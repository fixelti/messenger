package user

import (
	"fmt"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"message/internal/apperror"
	"message/internal/handlers"
	"message/pkg/logging"
	"net/http"
	"time"
)

const (
	userURL = "/users"
)

const (
	IdentityJWTKet = "id"
	RoleJWTKey     = "role"
)

type handler struct {
	logger     *logging.Logger
	repository Repository
}

type JwtWrapper struct {
	SecretKey       string
	Issuer          string
	ExpirationHours int64
}

type IDRequest struct {
	UserID uint `uri:"user_id" binding:"required,min=1"`
}

//TODO: Разобраться в этом блоке кода

func NewHandler(logger *logging.Logger, repository Repository) handlers.Handler {
	return &handler{
		logger:     logger,
		repository: repository,
	}
}

func (h *handler) Register(router *gin.RouterGroup) {
	jwtMiddleware := h.SignIn()
	users := router.Group(userURL)
	{
		users.POST("/signin", jwtMiddleware.LoginHandler)
		users.Use(jwtMiddleware.MiddlewareFunc())
		{
			users.POST("", apperror.Middleware(h.Create))
			users.DELETE("/:user_id", apperror.Middleware(h.Delete))
		}
		//...//
	}
}

func (h *handler) Create(c *gin.Context) error {
	var userDTO CreateUserDTO
	if err := c.ShouldBindJSON(&userDTO); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userDTO.Password), bcrypt.MinCost)
	if err != nil {
		return err
	}

	newUser, err := h.repository.Create(User{
		Email:      userDTO.Email,
		Login:      userDTO.Login,
		Password:   string(hashedPassword),
		SecretWord: userDTO.SecretWord,
		UserRole:   1,
	})
	if err != nil {
		return apperror.NewAppError(nil, "internal server error", "don't create user", "USR-0000001")
	}

	c.JSON(http.StatusOK, newUser)
	return nil
}

func (h *handler) Delete(c *gin.Context) error {
	var userID IDRequest
	if err := c.ShouldBindUri(&userID); err != nil {
		return err
	}

	err := h.repository.Delete(userID.UserID)
	if err != nil {
		return apperror.NewAppError(nil, "internal server error", "can't delete user", "USR-0000004")
	}
	c.IndentedJSON(http.StatusOK, "Deleted")
	return nil
}

func (h *handler) SignIn() *jwt.GinJWTMiddleware {
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
			if v, ok := data.(*User); ok {
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
				return &User{
					ID: v,
				}
			}
			return &User{
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

			queryUser, _ := h.repository.ReadByLogin(credentials.Login)
			if queryUser.ID == 0 {
				return "", jwt.ErrFailedAuthentication
			}

			err := bcrypt.CompareHashAndPassword([]byte(queryUser.Password), []byte(credentials.Password))
			if err != nil {
				return "", jwt.ErrFailedAuthentication
			}
			return &queryUser, nil
		},

		Authorizator: func(data interface{}, c *gin.Context) bool {
			if _, ok := data.(*User); ok {
				return true
			}
			fmt.Println(data.(*User))
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
		h.logger.Tracef("Can't wake up JWT Middleware! Error: %s\n", err.Error())
		return nil
	}

	errInit := m.MiddlewareInit()
	if errInit != nil {
		h.logger.Tracef("Can't init JWT Middleware! Error: %s\n", errInit.Error())
		return nil
	}

	return m
}
