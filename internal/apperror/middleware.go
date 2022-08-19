package apperror

import (
	"errors"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"message/internal/user"
	"message/pkg/client/postgresql"
	"message/pkg/logging"
	"net/http"
	"time"
)

const (
	IdentityJWTKet = "id"
	RoleJWTKey = "role"
)

type UserMiddleware struct {
	client postgresql.Client
	logger *logging.Logger
}

type JwtWrapper struct {
	SecretKey string
	Issuer string
	ExpirationHours int64
}

// TODO: разобраться во всехблоках кода
type appHandler func(ctx *gin.Context) error

func Middleware(handler appHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var appErr *AppError
		var ok bool

		err := handler(ctx)
		if err != nil {
			if errors.As(err, &appErr) {
				if errors.Is(err, ErrNotFound) {
					ctx.IndentedJSON(http.StatusNotFound, ErrNotFound)
					return
				}
				if errors.Is(err, ErrNotAuth) {
					ctx.IndentedJSON(http.StatusUnauthorized, ErrNotAuth)
					return
				}

				appErr, ok = err.(*AppError)
				if !ok {
					panic("can't convert err to app error")
				}
				ctx.IndentedJSON(http.StatusBadRequest, appErr)
				return
			}
			ctx.IndentedJSON(http.StatusTeapot, appErr.systemError(err))
		}
	}
}

func (u *UserMiddleware) JwtMiddleware() *jwt.GinJWTMiddleware {
	m, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm: "messenger",
		Key: []byte("test"), // TODO: сделать, чтобы ключ брался из конфигов
		Timeout: time.Minute * 100,
		MaxRefresh: time.Minute * 1000,
		IdentityKey: IdentityJWTKet,
		RefreshResponse: func (c *gin.Context, code int, token string, t time.Time) {

			c.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"token": token,
				"expire": t.Format(time.RFC3339),
				"message": "refresh successfully",
			})
		},

		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*user.User); ok {
				return jwt.MapClaims{
					IdentityJWTKet: v.ID,
					"role": v.Role
				}
			}
			return jwt.MapClaims{
				"error": true,
			}
		},

		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			if v, ok := claims[IdentityJWTKet].(uint); ok {
				return &user.User{
					ID: v,
				}
			}
			return &user.User{
				ID: 0,
			}
		},
	}
}