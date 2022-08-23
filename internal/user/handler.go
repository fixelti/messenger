package user

import (
	"fmt"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"message/internal/apperror"
	"message/internal/handlers"
	"message/internal/middleware"
	"message/pkg/client/postgresql"
	"message/pkg/logging"
	"message/pkg/service/checkingForBannedUser"
	"net/http"
)

const (
	userURL = "/users"
)

type handler struct {
	logger          *logging.Logger
	repository      Repository
	userMiddleware  middleware.UserMiddleware
	checkBannedUser checkingForBannedUser.CheckingBannedUser
}

type IDRequest struct {
	UserID uint `uri:"user_id" binding:"required,min=1"`
}

//TODO: Разобраться в этом блоке кода

func NewHandler(logger *logging.Logger, repository Repository, client postgresql.Client) handlers.Handler {
	return &handler{
		logger:          logger,
		repository:      repository,
		userMiddleware:  middleware.UserMiddleware{Client: client, Logger: logger},
		checkBannedUser: checkingForBannedUser.CheckingBannedUser{Client: client, Logger: logger},
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
			users.GET("/:user_id", apperror.Middleware(h.Read))
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
		UserRole:   3,
	})
	if err != nil {
		return apperror.NewAppError(nil, "internal server error", "don't create user", "USR-0000001")
	}

	c.JSON(http.StatusOK, newUser)
	return nil
}

func (h *handler) Read(c *gin.Context) error {
	var request IDRequest
	var user User
	var err error
	claims := jwt.ExtractClaims(c)
	userID := claims[middleware.IdentityJWTKet].(float64)
	userRole := claims[middleware.RoleJWTKey].(float64)
	if err := c.ShouldBindUri(&request); err != nil {
		return err
	}

	if uint(userRole) <= middleware.Admin {
		user, err = h.repository.Read(request.UserID)
		if err != nil {
			return apperror.NewAppError(nil, "user not found", "user not found", "USR-0000006")
		}

	} else if uint(userRole) == middleware.User {
		banned, err := h.checkBannedUser.CheckingBannedUser(uint(userID), request.UserID)
		fmt.Println("err: ", err)
		if err != nil {
			return apperror.NewAppError(nil, "server error", "internal server error", "USR-0000007")
		}

		if !banned {
			c.JSON(http.StatusForbidden, gin.H{"error": "you are banned from this user"})
			return nil
		}

		user, err = h.repository.Read(request.UserID)
		if err != nil {
			return apperror.NewAppError(nil, "user not found", "user not found", "USR-0000006")
		}
	}
	c.JSON(http.StatusOK, user)
	return nil
}

func (h *handler) Delete(c *gin.Context) error {
	var requestUserID IDRequest
	claims := jwt.ExtractClaims(c)
	userID := claims[middleware.IdentityJWTKet].(float64)
	userRole := claims[middleware.RoleJWTKey].(float64)
	if err := c.ShouldBindUri(&requestUserID); err != nil {
		return err
	}

	if _, err := h.repository.Read(requestUserID.UserID); err != nil {
		return apperror.NewAppError(nil, "not found", "user not found", "USR-0000005")
	}

	if uint(userRole) <= middleware.Admin {
		err := h.repository.Delete(requestUserID.UserID)
		if err != nil {
			return apperror.NewAppError(nil, "internal server error", "can't delete user", "USR-0000004")
		}
		c.JSON(http.StatusOK, "Deleted")

	} else if uint(userRole) == middleware.User && uint(userID) == requestUserID.UserID {
		err := h.repository.Delete(requestUserID.UserID)
		if err != nil {
			return apperror.NewAppError(nil, "internal server error", "can't delete user", "USR-0000004")
		}
		c.JSON(http.StatusOK, "Deleted")

	} else {
		return apperror.NewAppError(nil, "not your record", "can't delete user", "URS-0000005")
	}
	return nil
}
