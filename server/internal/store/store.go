// Пакет store. Хранилище
package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/iurnickita/gophkeeper1/server/internal/model"
	"github.com/iurnickita/gophkeeper1/server/internal/store/config"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Store интерфейс хранилище
type Store interface {
	AuthRegister(ctx context.Context, login string, password string) (int, error)
	AuthLogin(ctx context.Context, login string, password string) (int, error)
	List(ctx context.Context, userID int) ([]string, error)
	Read(ctx context.Context, userID int, unitName string) (model.Unit, error)
	Write(ctx context.Context, unit model.Unit) error
	Delete(ctx context.Context, userID int, unitName string) error
	GetEncryptSK(ctx context.Context) ([]string, error)
	SetEncryptSK(ctx context.Context, sk string) error
}

var (
	ErrNoRows        = errors.New("no rows")
	ErrAlreadyExists = errors.New("already exists")
)

// psqlStore postgresql реализация интерфейса хранилища
type psqlStore struct {
	database *sql.DB
}

// AuthRegister implements Store.
func (s *psqlStore) AuthRegister(ctx context.Context, login string, password string) (int, error) {
	// Запись нового пользователя
	row := s.database.QueryRowContext(ctx,
		"INSERT INTO auth (login, password)"+
			" VALUES ($1, $2)"+
			" RETURNING userid",
		login,
		password)

	// Получение ID пользователя
	var userid int
	err := row.Scan(&userid)
	if err != nil {
		// Проверка: уже существует
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return 0, ErrAlreadyExists
			}
		}

		return 0, err
	}

	return userid, nil
}

// AuthLogin implements Store.
func (s *psqlStore) AuthLogin(ctx context.Context, login string, password string) (int, error) {
	// Получение ID пользователя
	row := s.database.QueryRowContext(ctx,
		"SELECT userid FROM auth"+
			" WHERE login = $1",
		login)
	var userid int
	err := row.Scan(&userid)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrNoRows
		}
		return 0, err
	}

	return userid, nil
}

// List implements Store.
func (s *psqlStore) List(ctx context.Context, userID int) ([]string, error) {
	panic("unimplemented")
}

// Read implements Store.
func (s *psqlStore) Read(ctx context.Context, userID int, unitName string) (model.Unit, error) {
	row := s.database.QueryRowContext(ctx,
		"SELECT userid, unitname, uploadedat, type, datask, data"+
			" FROM data_units"+
			" WHERE userid   = $1"+
			"   AND unitname = $2",
		userID,
		unitName)
	var unit model.Unit
	err := row.Scan(&unit.Key.UserID,
		&unit.Key.UnitName,
		&unit.Meta.UploadedAt,
		&unit.Meta.Type,
		&unit.Meta.DataSK,
		&unit.Data)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Unit{}, ErrNoRows
		}
		return model.Unit{}, err
	}
	return unit, nil
}

// Write implements Store.
func (s *psqlStore) Write(ctx context.Context, unit model.Unit) error {
	_, err := s.database.ExecContext(ctx,
		"INSERT INTO data_units (userid, unitname, uploadedat, type, datask, data)"+
			"VALUES ($1, $2, $3, $4, $5, $6)",
		unit.Key.UserID,
		unit.Key.UnitName,
		time.Now(),
		unit.Meta.Type,
		unit.Meta.DataSK,
		unit.Data)
	if err != nil {
		// Проверка: уже существует
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ErrAlreadyExists
			}
		}
		return err
	}
	return nil
}

// Delete implements Store.
func (s *psqlStore) Delete(ctx context.Context, userID int, unitName string) error {
	panic("unimplemented")
}

// GetEncryptSK
func (s *psqlStore) GetEncryptSK(ctx context.Context) ([]string, error) {
	rows, err := s.database.QueryContext(ctx,
		"SELECT body"+
			" FROM encryption_sk")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Err() != nil {
		return nil, err
	}
	var keys []string
	for rows.Next() {
		var key string
		err := rows.Scan(&key)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	if rows.Err() != nil {
		return nil, err
	}

	return keys, nil
}

// SetEncryptSK
func (s *psqlStore) SetEncryptSK(ctx context.Context, sk string) error {
	_, err := s.database.ExecContext(ctx,
		"INSERT INTO encryption_sk (body)"+
			" VALUES ($1)",
		sk)
	if err != nil {
		return err
	}
	return nil
}

// NewStore создает объект хранилища
func NewStore(cfg config.Config) (Store, error) {
	db, err := sql.Open("pgx", cfg.DBDsn)
	if err != nil {
		return nil, err
	}

	// Таблица учетных записей
	_, err = db.Exec(
		"CREATE TABLE IF NOT EXISTS auth (" +
			" login VARCHAR (20) PRIMARY KEY," +
			" userid SERIAL UNIQUE," +
			" password VARCHAR (30) NOT NULL" +
			" );")
	if err != nil {
		return nil, err
	}

	// Таблица данных
	_, err = db.Exec(
		"CREATE TABLE IF NOT EXISTS data_units (" +
			" userid INTEGER," +
			" unitname VARCHAR (20) NOT NULL," +
			" uploadedat TIMESTAMP NOT NULL," +
			" type SMALLINT NOT NULL," +
			" datask VARCHAR (400) NOT NULL," +
			" data BYTEA NOT NULL," +
			" PRIMARY KEY (userid, unitname)" +
			" );")
	if err != nil {
		return nil, err
	}

	// Таблица промежуточных(постоянных) паролей
	_, err = db.Exec(
		"CREATE TABLE IF NOT EXISTS encryption_sk (" +
			" id SERIAL PRIMARY KEY," +
			" body VARCHAR (400)" +
			" );")
	if err != nil {
		return nil, err
	}

	return &psqlStore{
		database: db,
	}, nil
}
