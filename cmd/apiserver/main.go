package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"message/internal/config"
	"message/internal/user"
	userdb "message/internal/user/db"
	"message/pkg/client/postgresql"
	"message/pkg/logging"
)

func main() {
	appContext := context.Background()
	logger := logging.GetLogger()
	router := gin.Default()
	v1Group := router.Group("/api/v1")

	appConfig := config.GetConfig()
	pgConn := postgresql.NewClient(appContext, *appConfig)

	userRepo := userdb.NewRepository(pgConn, logger)
	userController := user.NewHandler(logger, userRepo)
	userController.Register(v1Group)

	logger.Fatalln(router.Run(":8081"))
}
