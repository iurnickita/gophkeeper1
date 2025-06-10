package auth

import (
	"context"
	"strconv"

	"github.com/iurnickita/gophkeeper1/internal/server/store"
	"github.com/iurnickita/gophkeeper1/internal/server/token"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Register(ctx context.Context, login string, password string) (string, error)
	Login(ctx context.Context, login string, password string) (string, error)
	AuthUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)
}

type Key string

const (
	metadataUserToken string = "token"
	ContextUserID     Key    = "userID"
)

type auth struct {
	store store.Store
}

func NewAuth(store store.Store) (Auth, error) {
	return &auth{store: store}, nil
}

// Register
// Ошибки: store.ErrAlreadyExists
func (a *auth) Register(ctx context.Context, login string, password string) (string, error) {
	// Запись в БД
	userID, err := a.store.AuthRegister(ctx, login, password)
	if err != nil {
		return "", err
	}

	// Запись ID в JWT-токен
	tokenString, err := token.BuildJWTString(strconv.Itoa(userID))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Login
// Ошибки: store.ErrNoRows
func (a *auth) Login(ctx context.Context, login string, password string) (string, error) {
	// Проверка в БД
	userID, err := a.store.AuthLogin(ctx, login, password)
	if err != nil {
		return "", err
	}

	// Запись ID в JWT-токен
	tokenString, err := token.BuildJWTString(strconv.Itoa(userID))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// AuthUnaryInterceptor прослойка аутентификации для gRPC хендлеров
func (a *auth) AuthUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Обход для регистрации/входа
	if info.FullMethod == "/gophkeeper.Gophkeeper/Register" || info.FullMethod == "/gophkeeper.Gophkeeper/Authenticate" {
		return handler(ctx, req)
	}

	// Получение метаданных из контекста
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		var t string
		var userID string
		var err error
		// Чтение токена из метаданных
		values := md.Get(metadataUserToken)
		if len(values) > 0 {
			// Получение кода пользователя из токена
			t = values[0]
			userID, err = token.GetUserID(t)
			if err != nil {
				return nil, status.Errorf(codes.Unauthenticated, err.Error())
			}
		} else {
			return nil, status.Errorf(codes.Unauthenticated, "%s Unauthenticated. Use Register procedure", info.FullMethod)
		}
		// Запись кода пользователя в контекст для дальнейшего использования
		ctx = context.WithValue(ctx, ContextUserID, userID)
	}

	return handler(ctx, req)
}
