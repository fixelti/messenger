package user

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"message/internal/apperror"
	"message/internal/handlers"
	"message/internal/middleware"
	"message/pkg/client/postgresql"
	"message/pkg/logging"
	"net/http"
)

const (
	userURL = "/users"
)

type handler struct {
	logger         *logging.Logger
	repository     Repository
	userMiddleware middleware.UserMiddleware
}

type IDRequest struct {
	UserID uint `uri:"user_id" binding:"required,min=1"`
}

//TODO: Разобраться в этом блоке кода

func NewHandler(logger *logging.Logger, repository Repository, client postgresql.Client) handlers.Handler {
	return &handler{
		logger:         logger,
		repository:     repository,
		userMiddleware: middleware.UserMiddleware{Client: client, Logger: logger},
	}
}

func (h *handler) Register(router *gin.RouterGroup) {
	jwtMiddleware := h.userMiddleware.JwtMiddleware()
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
