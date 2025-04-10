package handlers

import (
	"Datapolis/internal/models"
	"Datapolis/internal/services"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	log.Printf("Попытка регистрации пользователя: %s", user.Username)

	err := h.userService.Register(c.Request.Context(), &user)
	if err != nil {
		switch err {
		case service.ErrUserExists:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case service.ErrInvalidUserData:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			log.Printf("Неизвестная ошибка при регистрации пользователя: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		}
		return
	}

	user.Password = ""
	log.Printf("Успешная регистрация пользователя: %s", user.Username)
	c.JSON(http.StatusOK, gin.H{"message": "Пользователь успешно зарегистрирован"})

}

func (h *UserHandler) Login(c *gin.Context) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	user, err := h.userService.Login(c.Request.Context(), credentials.Username, credentials.Password)
	if err != nil {
		if err == service.ErrInvalidLogin {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверное имя пользователя или пароль"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}
