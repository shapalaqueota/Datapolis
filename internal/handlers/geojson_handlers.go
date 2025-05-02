package handlers

import (
	"Datapolis/internal/models"
	service "Datapolis/internal/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type GeoJSONHandler struct {
	geoJSONService *service.GeoService
}

// конструктор
func NewGeoJSONHandler(geoJSONService *service.GeoService) *GeoJSONHandler {
	return &GeoJSONHandler{geoJSONService: geoJSONService}
}

// GetCollection получает коллекцию по ID
func (h *GeoJSONHandler) GetCollection(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	collection, err := h.geoJSONService.GetCollection(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении коллекции: " + err.Error()})
		return
	}

	if collection == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Коллекция не найдена"})
		return
	}

	c.JSON(http.StatusOK, collection)
}

// ExportGeoJSON экспортирует коллекцию в GeoJSON формат
func (h *GeoJSONHandler) ExportGeoJSON(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	geoJSON, err := h.geoJSONService.ExportGeoJSON(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при экспорте GeoJSON: " + err.Error()})
		return
	}

	c.Header("Content-Type", "application/json")
	c.Writer.Write(geoJSON)
}

// DeleteCollection удаляет коллекцию
func (h *GeoJSONHandler) DeleteCollection(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	userID, _ := c.Get("user_id") // Исправлено с "userID" на "user_id"
	err = h.geoJSONService.DeleteCollection(c.Request.Context(), id, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении коллекции: " + err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetFeatures получает все фичи коллекции
func (h *GeoJSONHandler) GetFeatures(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	features, err := h.geoJSONService.GetFeatures(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении фич: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, features)
}

// AddFeature добавляет новую фичу в коллекцию
func (h *GeoJSONHandler) AddSingleFeature(c *gin.Context) {
	// Получаем ID коллекции из URL
	cid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный ID коллекции"})
		return
	}

	var feature models.GeoJSONFeature
	if err := c.ShouldBindJSON(&feature); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных: " + err.Error()})
		return
	}
	feature.CollectionID = cid

	if err := h.geoJSONService.AddSingleFeature(c.Request.Context(), &feature); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при добавлении фичи: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, feature)
}

// UpdateFeature обновляет фичу
func (h *GeoJSONHandler) UpdateFeature(c *gin.Context) {
	// 1) Парсим ID фичи из URL
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID фичи"})
		return
	}

	var input models.GeoJSONFeature
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных: " + err.Error()})
		return
	}
	input.ID = id

	existing, err := h.geoJSONService.GetFeatureByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка поиска фичи: " + err.Error()})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Фича не найдена"})
		return
	}
	input.CollectionID = existing.CollectionID

	// 4) Выполняем обновление
	if err := h.geoJSONService.UpdateFeature(c.Request.Context(), &input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении фичи: " + err.Error()})
		return
	}

	// 5) Отдаём обновлённую фичу (можно вернуть input или снова подгрузить из БД)
	c.JSON(http.StatusOK, input)
}

// DeleteFeature удаляет фичу
func (h *GeoJSONHandler) DeleteFeature(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	err = h.geoJSONService.DeleteFeature(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении фичи: " + err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetAllCollections получает все коллекции
func (h *GeoJSONHandler) GetAllCollections(c *gin.Context) {
	collections, err := h.geoJSONService.GetAllCollections(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении коллекций: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, collections)
}

func (h *GeoJSONHandler) UploadGeoJSONBulk(c *gin.Context) {
	// 1) достаём файл
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл не найден: " + err.Error()})
		return
	}
	defer file.Close()

	// 2) читаем метаданные
	name := c.PostForm("name")
	if name == "" {
		name = "unnamed_collection"
	}
	description := c.PostForm("description")

	// 3) получаем user_id из контекста (middleware)
	uidIfc, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неавторизованный запрос"})
		return
	}
	userID, ok := uidIfc.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Неверный формат user_id"})
		return
	}

	// 4) вызываем сервис bulk‑импорта
	col, err := h.geoJSONService.ImportGeoJSONBulk(
		c.Request.Context(),
		file,
		name,
		description,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка bulk‑импорта: " + err.Error()})
		return
	}

	// 5) возвращаем созданную коллекцию (ID, timestamps)
	c.JSON(http.StatusCreated, col)
}
