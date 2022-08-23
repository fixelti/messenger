package middleware

import (
	"context"
	"errors"
	"fmt"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgconn"
	"golang.org/x/crypto/bcrypt"
	"message/pkg/client/postgresql"
	"message/pkg/logging"
	"net/http"
	"time"
)

type Test struct {
	ID         uint `json:"id"`
	CreatedAt  time.Time
	DeletedAt  *time.Time `sql:"index"`
	Email      string     `json:"email"`
	Login      string     `json:"login"`
	Password   string     `json:"password"`
	SecretWord string     `json:"secret_word"`
	FindVision bool       `json:"find_vision"`
	AddFriend  bool       `json:"add_friend"`
	Friends    []uint     `json:"friends"`
	UserRole   uint       `json:"user_role"`
}

const (
	Root  = 1
	Admin = 2
	User  = 3
)

const (
	IdentityJWTKet = "id"
	RoleJWTKey     = "role"
)

type JwtWrapper struct {
	SecretKey       string
	Issuer          string
	ExpirationHours int64
}

type UserMiddleware struct {
	Client postgresql.Client
	Logger *logging.Logger
}

func (u *UserMiddleware) JwtMiddleware() *jwt.GinJWTMiddleware {
	m, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "messenger",
		Key:         []byte("test"), // TODO: сделать, чтобы ключ брался из конфигов
		Timeout:     time.Minute * 100,
		MaxRefresh:  time.Minute * 1000,
		IdentityKey: IdentityJWTKet,
		RefreshResponse: func(c *gin.Context, code int, token string, t time.Time) {

			c.JSON(http.StatusOK, gin.H{
				"code":    http.StatusOK,
				"token":   token,
				"expire":  t.Format(time.RFC3339),
				"message": "refresh successfully",
			})
		},

		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(Test); ok {
				return jwt.MapClaims{
					IdentityJWTKet: v.ID,
					"role":         v.UserRole,
				}
			}
			return jwt.MapClaims{
				"error": true,
			}
		},

		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			if v, ok := claims[IdentityJWTKet].(uint); ok {
				return &Test{
					ID: v,
				}
			}
			return &Test{
				ID: 0,
			}
		},

		Authenticator: func(c *gin.Context) (interface{}, error) {
			var credentials = struct {
				Login    string `form:"login" json:"login" binding:"required"`
				Password string `form:"password" json:"password" binding:"required"`
			}{}

			if err := c.ShouldBind(&credentials); err != nil {
				return "", jwt.ErrMissingLoginValues
			}

			var queryUser []*Test

			request := `SELECT * FROM users WHERE login = $1;`

			tx, err := u.Client.Begin(context.Background())
			if err != nil {
				_ = tx.Rollback(context.Background())
				u.Logger.Tracef("can't start transaction: %s", err.Error())
				return nil, err
			}

			err = pgxscan.Select(context.Background(), u.Client, &queryUser, request, credentials.Login)
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
					u.Logger.Error(newErr)
					return nil, newErr
				}
				u.Logger.Error(err)
				return nil, err
			}
			_ = tx.Commit(context.Background())

			err = bcrypt.CompareHashAndPassword([]byte(queryUser[0].Password), []byte(credentials.Password))
			if err != nil {
				return "", jwt.ErrFailedAuthentication
			}
			return *queryUser[0], nil
		},

		Authorizator: func(data interface{}, c *gin.Context) bool {
			if _, ok := data.(*Test); ok {
				return true
			}
			return false
		},

		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		TokenHeadName:     "Bearer ",
		TokenLookup:       "header: Authorization, query: token, cookie: jwt",
		TimeFunc:          time.Now,
		SendAuthorization: true,
	},
	)

	if err != nil {
		u.Logger.Tracef("Can't wake up JWT Middleware! Error: %s\n", err.Error())
		return nil
	}

	errInit := m.MiddlewareInit()
	if errInit != nil {
		u.Logger.Tracef("Can't init JWT Middleware! Error: %s\n", errInit.Error())
		return nil
	}

	return m
}

func (u *UserMiddleware) RootMiddleware(c *gin.Context) {
	var user Test
	var err error
	claims := jwt.ExtractClaims(c)
	userID := claims[IdentityJWTKet].(float64)

	user, err = u.findUser(uint(userID))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}

	if user.UserRole != Root {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient rights"})
		return
	}

	c.Next()
}

func (u *UserMiddleware) AdminMiddleware(c *gin.Context) {
	var user Test
	var err error
	claims := jwt.ExtractClaims(c)
	userID := claims[IdentityJWTKet].(float64)

	user, err = u.findUser(uint(userID))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}

	if user.UserRole != Root {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient rights"})
		return
	}

	c.Next()
}

func (u *UserMiddleware) findUser(userID uint) (Test, error) {
	var queryUser []*Test
	request := `SELECT * FROM users WHERE id = $1;`

	tx, err := u.Client.Begin(context.Background())
	if err != nil {
		_ = tx.Rollback(context.Background())
		u.Logger.Tracef("can't start transaction: %s", err.Error())
		return Test{}, err
	}

	err = pgxscan.Select(context.Background(), u.Client, &queryUser, request, userID)
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
			u.Logger.Error(newErr)
			return Test{}, newErr
		}
		u.Logger.Error(err)
		return Test{}, err
	}
	_ = tx.Commit(context.Background())
	return *queryUser[0], nil
}
