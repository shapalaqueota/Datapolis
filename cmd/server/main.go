package main

import (
	"Datapolis/internal/db"
	"Datapolis/internal/handlers"
	"Datapolis/internal/repository"
	"Datapolis/internal/routes"
	service "Datapolis/internal/services"
	"log"
)

func main() {
	db.ConnectDB()
	defer db.Pool.Close()

	userRepo := repository.NewUserRepository(db.Pool)
	userService := service.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)
	authService := service.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)

	router := routes.Router(userHandler, authHandler)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
