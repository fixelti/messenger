package apperror

import (
	"errors"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"message/internal/user"
	"message/pkg/client/postgresql"
	"message/pkg/logging"
	"net/http"
	"time"
)

const (
	IdentityJWTKet = "id"
	RoleJWTKey     = "role"
)

type UserMiddleware struct {
	client     postgresql.Client
	logger     *logging.Logger
	repository user.Repository
}

type JwtWrapper struct {
	SecretKey       string
	Issuer          string
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
			if v, ok := data.(*user.User); ok {
				return jwt.MapClaims{
					IdentityJWTKet: v.ID,
					"role":         v.Role,
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

		Authenticator: func(c *gin.Context) (interface{}, error) {
			var credentials = struct {
				Login    string `form:"login" json:"login" binding:"required"`
				Password string `form:"password" json:"password" binding:"required"`
			}{}

			if err := c.ShouldBind(&credentials); err != nil {
				return "", jwt.ErrMissingLoginValues
			}

			var userModel user.User
			queryUser, _ := u.repository.ReadByLogin(credentials.Login)
			if queryUser.ID == 0 {
				return "", jwt.ErrFailedAuthentication
			}

			err := bcrypt.CompareHashAndPassword([]byte(userModel.Password), []byte(credentials.Password))
			if err != nil {
				return "", jwt.ErrFailedAuthentication
			}

			return &userModel, nil
		},

		Authorizator: func(data interface{}, c *gin.Context) bool {
			if _, ok := data.(*user.User); ok {
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
		u.logger.Tracef("Can't wake up JWT Middleware! Error: %s\n", err.Error())
		return nil
	}

	errInit := m.MiddlewareInit()
	if errInit != nil {
		u.logger.Tracef("Can't init JWT Middleware! Error: %s\n", errInit.Error())
		return nil
	}

	return m
}
