package user

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"message/internal/apperror"
	"message/internal/handlers"
	"message/pkg/logging"
	"net/http"
)

const (
	userURL = "/users"
)

type handler struct {
	logger     *logging.Logger
	repository Repository
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
	users := router.Group(userURL)
	{
		users.POST("", apperror.Middleware(h.Create))
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
	})
	if err != nil {
		return apperror.NewAppError(nil, "new app error", "new app error", "NEW-0000001")
	}

	c.JSON(http.StatusOK, newUser)
	return nil
}
