// Пакет grpcclient
package grpcclient

import (
	"context"

	"github.com/iurnickita/gophkeeper1/internal/client/grpc_client/client/config"
	"github.com/iurnickita/gophkeeper1/internal/client/model"
	gophTLS "github.com/iurnickita/gophkeeper1/internal/client/tls"
	pb "github.com/iurnickita/gophkeeper1/internal/contract/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Client
type Client struct {
	conn       *grpc.ClientConn
	gophkeeper pb.GophkeeperClient
}

// Register
func (c Client) Register(login string, password string) (string, error) {
	ctx := context.Background()
	req := pb.RegisterRequest{Login: login, Password: password}
	resp, err := c.gophkeeper.Register(ctx, &req)
	if err != nil {
		return "", err
	}

	return resp.Token, nil
}

// Authenticate
func (c Client) Authenticate(login string, password string) (string, error) {
	ctx := context.Background()
	req := pb.AuthenticateRequest{Login: login, Password: password}
	resp, err := c.gophkeeper.Authenticate(ctx, &req)
	if err != nil {
		return "", err
	}

	return resp.Token, nil
}

// List возвращает список единиц данных, хранящихся на сервере
func (c Client) List(token string) ([]string, error) {
	ctx := c.createContext(token)

	// Запрос
	resp, err := c.gophkeeper.List(ctx, &pb.Empty{})
	if err != nil {
		return nil, err
	}

	return resp.Unitname, nil
}

// Read
func (c Client) Read(token string, unitname string) (model.Unit, error) {
	ctx := c.createContext(token)

	// Запрос
	resp, err := c.gophkeeper.Read(ctx, &pb.ReadRequest{Unitname: unitname})
	if err != nil {
		return model.Unit{}, err
	}

	// Маппинг
	var unit model.Unit
	unit.Name = unitname
	unit.Body.Meta.Type = int(resp.Unittype)
	unit.Body.Data = resp.Unitdata

	return unit, nil
}

// Write
func (c Client) Write(token string, unit model.Unit) error {
	ctx := c.createContext(token)

	// Запрос
	req := &pb.WriteRequest{Unitname: unit.Name,
		Unittype: int32(unit.Body.Meta.Type),
		Unitdata: unit.Body.Data}
	_, err := c.gophkeeper.Write(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// Delete
func (c Client) Delete(token string, unitname string) error {
	ctx := c.createContext(token)

	// Запрос
	_, err := c.gophkeeper.Delete(ctx, &pb.DeleteRequest{Unitname: unitname})
	if err != nil {
		return err
	}

	return nil
}

// createContext создает контекст с метаданными для запроса к grpc-серверу
func (c Client) createContext(token string) context.Context {
	ctx := context.Background()
	// Передача токена в метаданных
	md := metadata.Pairs("token", token)
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx
}

func (c Client) Close() {
	c.conn.Close()
}

// NewClient создает grpc-клиент
func NewClient(cfg config.Config) (Client, error) {
	tlsCredentials, err := gophTLS.LoadTLSCredentials()
	if err != nil {
		return Client{}, err
	}

	// устанавливаем соединение с сервером
	conn, err := grpc.NewClient(":3200", grpc.WithTransportCredentials(tlsCredentials))
	if err != nil {
		return Client{}, err
	}
	// получаем переменную интерфейсного типа ShortenerClient,
	// через которую будем отправлять сообщения
	c := pb.NewGophkeeperClient(conn)

	return Client{conn: conn, gophkeeper: c}, nil
}
