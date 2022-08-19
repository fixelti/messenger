package userdb

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"message/internal/user"
	"message/pkg/client/postgresql"
	"message/pkg/logging"
)

type repository struct {
	client postgresql.Client
	logger *logging.Logger
}

func (r *repository) Create(newUser user.User) (user.User, error) {
	request := `
		INSERT INTO users(email, login, password, secret_word) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id
		`

	tx, err := r.client.Begin(context.Background())
	if err != nil {
		_ = tx.Rollback(context.Background())
		r.logger.Tracef("can't start transaction: %s", err.Error()) // Прочитать про Tracef
		return user.User{}, err
	}

	err = tx.QueryRow(
		context.Background(),
		request,
		newUser.Email,
		newUser.Login,
		newUser.Password,
		newUser.SecretWord).Scan(&newUser.ID)

	if err != nil {
		_ = tx.Rollback(context.Background())
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			pgErr = err.(*pgconn.PgError) // Прочитать про то, почему тут точка
			newErr := fmt.Errorf("SQL Error: %s, Detail: %s, Where %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
			r.logger.Error(newErr)
			return user.User{}, newErr
		}
		r.logger.Error(err)
		return user.User{}, err
	}
	_ = tx.Commit(context.Background())
	return newUser, nil
}

func (r *repository) Read(userId uint) (user.User, error) { return user.User{}, nil }

func (r *repository) ReadByLogin(login string) (user.User, error) {
	var queryUser user.User

	request := `SELECT * FROM users WHERE login = $1;`

	tx, err := r.client.Begin(context.Background())
	if err != nil {
		_ = tx.Rollback(context.Background())
		r.logger.Tracef("can't start transaction: %s", err.Error())
		return user.User{}, err
	}

	err = tx.QueryRow(context.Background(), request, login).Scan(queryUser)
	if err != nil {
		_ = tx.Rollback(context.Background())
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			pgErr = err.(*pgconn.PgError)
			newErr := fmt.Errorf(
				"SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
				pgErr.Message,
				pgErr.Detail,
				pgErr.Where,
				pgErr.Code,
				pgErr.SQLState(),
			)
			r.logger.Error(newErr)
			return user.User{}, newErr
		}
		r.logger.Error(err)
		return user.User{}, err
	}
	_ = tx.Commit(context.Background())
	return queryUser, nil
}

func (r *repository) List(filter user.Filter) (user.Pagination, error) { return user.Pagination{}, nil }

func (r *repository) Update(userId uint, userToUpdate user.User) (user.User, error) {
	return user.User{}, nil
}

func (r *repository) Delete(userId uint) error { return nil }

func NewRepository(client postgresql.Client, logger *logging.Logger) user.Repository {
	return &repository{
		client: client,
		logger: logger,
	}
}
