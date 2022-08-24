package userdb

import (
	"context"
	"errors"
	"fmt"
	"github.com/georgysavva/scany/pgxscan"
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
		INSERT INTO users(email, login, password, secret_word, user_role) 
		VALUES ($1, $2, $3, $4, $5) 
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
		newUser.SecretWord,
		newUser.UserRole).Scan(&newUser.ID)

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

func (r *repository) Read(userID uint) (user.User, error) {
	var queryUser user.User
	request := `SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL`

	tx, err := r.client.Begin(context.Background())
	if err != nil {
		_ = tx.Rollback(context.Background())
		r.logger.Tracef("can't start transaction: %s", err.Error())
		return user.User{}, err
	}

	err = pgxscan.Get(context.Background(), r.client, &queryUser, request, userID)
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
	return queryUser, err
}

func (r *repository) ReadByLogin(login string) (user.User, error) {
	var queryUser []*user.User

	request := `SELECT * FROM users WHERE login = $1;`

	tx, err := r.client.Begin(context.Background())
	if err != nil {
		_ = tx.Rollback(context.Background())
		r.logger.Tracef("can't start transaction: %s", err.Error())
		return user.User{}, err
	}

	err = pgxscan.Select(context.Background(), r.client, &queryUser, request, login)
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
	return *queryUser[0], nil
}

func (r *repository) List(filter user.Filter) (user.Pagination, error) { return user.Pagination{}, nil }

func (r *repository) Update(userToUpdate user.User) (user.User, error) {
	request := `
			UPDATE users
			SET find_vision = $1,
			add_friend = $2
			WHERE id = $3;`

	tx, err := r.client.Begin(context.Background())
	if err != nil {
		_ = tx.Rollback(context.Background())
		r.logger.Tracef("can't start transaction: %s", err.Error())
		return user.User{}, err
	}

	_, err = tx.Exec(context.Background(),
		request,
		userToUpdate.FindVision,
		userToUpdate.AddFriend,
		userToUpdate.ID)
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
	return userToUpdate, nil
}

func (r *repository) Delete(userId uint) error {
	requestDelete := `UPDATE users SET deleted_at = current_timestamp WHERE id = $1 AND deleted_at IS NULL RETURNING id`

	tx, err := r.client.Begin(context.Background())
	if err != nil {
		_ = tx.Rollback(context.Background())
		r.logger.Tracef("can't start transaction: %s", err.Error())
		return err
	}

	_, err = tx.Exec(context.Background(), requestDelete, userId)
	if err != nil {
		_ = tx.Rollback(context.Background())
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			pgErr = err.(*pgconn.PgError)
			newErr := fmt.Errorf("SQL Error: %s, Detail: %s, Where %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
			r.logger.Error(newErr)
			return newErr
		}
		r.logger.Error(err)
		return err
	}
	_ = tx.Commit(context.Background())
	return nil
}

func NewRepository(client postgresql.Client, logger *logging.Logger) user.Repository {
	return &repository{
		client: client,
		logger: logger,
	}
}
