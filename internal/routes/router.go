package routes

import (
	"Datapolis/internal/handlers"
	"github.com/gin-gonic/gin"
)

func Router(userHandler *handlers.UserHandler) *gin.Engine {
	router := gin.Default()

	router.POST("/register", userHandler.Register)
	router.POST("/login", userHandler.Login)

	return router
}
