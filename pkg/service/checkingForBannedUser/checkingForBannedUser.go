package checkingForBannedUser

import (
	"context"
	"errors"
	"fmt"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgconn"
	"message/pkg/client/postgresql"
	"message/pkg/logging"
	"time"
)

type CheckingBannedUser struct {
	Client postgresql.Client
	Logger *logging.Logger
}

type bannedUser struct {
	ID           uint `json:"id"`
	CreatedAt    time.Time
	DeletedAt    *time.Time `sql:"index"`
	UserID       uint       `json:"user_id"`
	BannedUserID uint       `json:"banned_user_id"`
}

func (c *CheckingBannedUser) CheckingBannedUser(userID uint, banUserID uint) (bool, error) {
	var bannedUser bannedUser
	request := `SELECT * FROM users_banned WHERE user_id = $1 AND banned_user_id = $2`

	tx, err := c.Client.Begin(context.Background())
	if err != nil {
		_ = tx.Rollback(context.Background())
		c.Logger.Tracef("can't start transaction: %s", err.Error())
		return false, err
	}

	err = pgxscan.Get(context.Background(), c.Client, &bannedUser, request, userID, banUserID)
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
			c.Logger.Error(newErr)
			return false, newErr
		}
		c.Logger.Error(err)
		return false, err
	}
	_ = tx.Commit(context.Background())
	if bannedUser.ID == 0 {
		return false, nil
	}
	return true, nil
}
