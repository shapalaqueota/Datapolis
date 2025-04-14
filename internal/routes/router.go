package routes

import (
	"Datapolis/internal/handlers"
	"Datapolis/internal/middleware"
	"github.com/gin-gonic/gin"
)

func Router(userHandler *handlers.UserHandler, authHandler *handlers.AuthHandler) *gin.Engine {
	router := gin.Default()

	router.POST("/login", authHandler.Login)
	router.POST("/refresh", authHandler.RefreshToken)

	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/renovation")
	}

	admin := protected.Group("/admin")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		admin.POST("/register", userHandler.Register)
	}

	return router
}
