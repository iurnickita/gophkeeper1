// Пакет grpcserver. Обработчики grpc
package grpcserver

import (
	"context"
	"net"
	"strconv"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/iurnickita/gophkeeper1/internal/contract/proto"
	"github.com/iurnickita/gophkeeper1/internal/server/auth"
	gophTLS "github.com/iurnickita/gophkeeper1/internal/server/crypto/tls"
	"github.com/iurnickita/gophkeeper1/internal/server/grpc_server/server/config"
	"github.com/iurnickita/gophkeeper1/internal/server/model"
	"github.com/iurnickita/gophkeeper1/internal/server/service"
	"github.com/iurnickita/gophkeeper1/internal/server/store"
)

// Server grpc-обработчик
type Server struct {
	// нужно встраивать тип pb.Unimplemented<TypeName>
	// для совместимости с будущими версиями
	pb.UnimplementedGophkeeperServer

	config     config.Config
	auth       auth.Auth
	gophkeeper service.Service
	zaplog     *zap.Logger
	//wg         sync.WaitGroup
}

// NewServer создает новый grpc сервер
func NewServer(config config.Config, auth auth.Auth, gophkeeper service.Service, zaplog *zap.Logger) *Server {
	return &Server{
		config:     config,
		auth:       auth,
		gophkeeper: gophkeeper,
		zaplog:     zaplog,
	}
}

// Register
func (s *Server) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	token, err := s.auth.Register(ctx, in.Login, in.Password)
	if err != nil {
		switch err {
		case store.ErrAlreadyExists:
			return &pb.RegisterResponse{}, status.Error(codes.AlreadyExists, err.Error())
		// обработка неверного ввода: пустой логин, простой пароль
		// case :
		default:
			return &pb.RegisterResponse{}, status.Error(codes.Internal, err.Error())
		}
	}
	s.zaplog.Sugar().Debugf("register returns token: %s", token)
	return &pb.RegisterResponse{Token: token}, nil
}

// Authenticate
func (s *Server) Authenticate(ctx context.Context, in *pb.AuthenticateRequest) (*pb.AuthenticateResponse, error) {
	token, err := s.auth.Login(ctx, in.Login, in.Password)
	if err != nil {
		switch err {
		case store.ErrNoRows:
			return &pb.AuthenticateResponse{}, status.Error(codes.NotFound, err.Error())
		default:
			return &pb.AuthenticateResponse{}, status.Error(codes.Internal, err.Error())
		}
	}

	s.zaplog.Sugar().Debugf("authenticate returns token: %s", token)
	return &pb.AuthenticateResponse{Token: token}, nil
}

// List
func (s *Server) List(ctx context.Context, in *pb.Empty) (*pb.ListResponse, error) {
	// Код пользователя
	userID, err := strconv.Atoi(ctx.Value(auth.ContextUserID).(string))
	if err != nil {
		return &pb.ListResponse{}, status.Error(codes.Internal, err.Error())
	}

	// Получение списка доступных данных
	list, err := s.gophkeeper.List(ctx, userID)
	if err != nil {
		return &pb.ListResponse{}, status.Error(codes.Internal, err.Error())
	}
	return &pb.ListResponse{Unitname: list}, nil
}

// Read
func (s *Server) Read(ctx context.Context, in *pb.ReadRequest) (*pb.ReadResponse, error) {
	// Код пользователя
	userID, err := strconv.Atoi(ctx.Value(auth.ContextUserID).(string))
	if err != nil {
		return &pb.ReadResponse{}, status.Error(codes.Internal, err.Error())
	}

	// Чтение единицы данных
	unit, err := s.gophkeeper.Read(ctx, userID, in.Unitname)
	if err != nil {
		switch err {
		case store.ErrNoRows:
			return &pb.ReadResponse{}, status.Error(codes.NotFound, err.Error())
		default:
			return &pb.ReadResponse{}, status.Error(codes.Internal, err.Error())
		}
	}
	return &pb.ReadResponse{Unittype: int32(unit.Meta.Type), Unitdata: unit.Data}, nil
}

// Write
func (s *Server) Write(ctx context.Context, in *pb.WriteRequest) (*pb.Empty, error) {
	// Код пользователя
	userID, err := strconv.Atoi(ctx.Value(auth.ContextUserID).(string))
	if err != nil {
		return &pb.Empty{}, status.Error(codes.Internal, err.Error())
	}

	// Запись новой единицы данных
	var unit model.Unit
	unit.Key = model.UnitKey{UserID: userID, UnitName: in.Unitname}
	unit.Meta = model.UnitMeta{Type: int(in.Unittype)}
	unit.Data = in.Unitdata
	err = s.gophkeeper.Write(ctx, unit)
	if err != nil {
		switch err {
		case store.ErrAlreadyExists:
			return &pb.Empty{}, status.Error(codes.AlreadyExists, err.Error())
		default:
			return &pb.Empty{}, status.Error(codes.Internal, err.Error())
		}
	}
	return &pb.Empty{}, err
}

// Delete
func (s *Server) Delete(ctx context.Context, in *pb.DeleteRequest) (*pb.Empty, error) {
	// Код пользователя
	userID, err := strconv.Atoi(ctx.Value(auth.ContextUserID).(string))
	if err != nil {
		return &pb.Empty{}, status.Error(codes.Internal, err.Error())
	}

	// Чтение единицы данных
	err = s.gophkeeper.Delete(ctx, userID, in.Unitname)
	if err != nil {
		switch err {
		case store.ErrNoRows:
			return &pb.Empty{}, status.Error(codes.NotFound, err.Error())
		default:
			return &pb.Empty{}, status.Error(codes.Internal, err.Error())
		}
	}
	return &pb.Empty{}, nil
}

// Serve - запуск сервера
func Serve(cfg config.Config, auth auth.Auth, gophkeeper service.Service, zaplog *zap.Logger) error {
	// определяем порт для сервера
	listen, err := net.Listen("tcp", ":3200")
	if err != nil {
		return err
	}
	// tls
	tlsCredentials, err := gophTLS.LoadTLSCredentials()
	if err != nil {
		return err
	}

	// создаём gRPC-сервер
	s := grpc.NewServer(
		grpc.Creds(tlsCredentials),
		grpc.UnaryInterceptor(auth.AuthUnaryInterceptor))
	// создание обработчика
	h := NewServer(cfg, auth, gophkeeper, zaplog)
	// регистрируем сервис
	pb.RegisterGophkeeperServer(s, h)

	zaplog.Info("Сервер gRPC начал работу")
	// получаем запрос gRPC
	if err := s.Serve(listen); err != nil {
		return err
	}
	return nil
}
