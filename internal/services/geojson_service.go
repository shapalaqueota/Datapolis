package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"

	"Datapolis/internal/models"
	"Datapolis/internal/repository"
)

const srid4326 = 4326

type GeoService struct {
	repo *repository.GeoRepository
}

func NewGeoService(r *repository.GeoRepository) *GeoService { return &GeoService{repo: r} }

// GetCollection получает коллекцию по ID
func (s *GeoService) GetCollection(ctx context.Context, id int) (*models.GeoJSONCollection, error) {
	return s.repo.GetCollectionByID(ctx, id)
}

func (s *GeoService) ExportGeoJSON(ctx context.Context, collectionID int) ([]byte, error) {
	col, err := s.repo.GetCollectionByID(ctx, collectionID)
	if err != nil {
		return nil, err
	}
	if col == nil {
		return nil, errors.New("collection not found")
	}

	feats, err := s.repo.FeaturesByCollection(ctx, collectionID)
	if err != nil {
		return nil, err
	}

	out := map[string]any{
		"type": "FeatureCollection",
		"name": col.Name,
		"crs": map[string]any{
			"type":       "name",
			"properties": map[string]any{"name": "EPSG:" + strconv.Itoa(col.SRID)},
		},
	}
	flist := make([]map[string]any, 0, len(feats))
	for _, f := range feats {
		flist = append(flist, map[string]any{
			"type":       "Feature",
			"properties": json.RawMessage(f.Properties),
			"geometry":   json.RawMessage(f.Geometry),
		})
	}
	out["features"] = flist
	return json.Marshal(out)
}

// DeleteCollection удаляет коллекцию
func (s *GeoService) DeleteCollection(ctx context.Context, collectionID, userID int) error {
	return s.repo.DeleteCollection(ctx, collectionID, userID)
}

// GetFeatures получает все фичи коллекции
func (s *GeoService) GetFeatures(ctx context.Context, collectionID int) ([]*models.GeoJSONFeature, error) {
	return s.repo.GetFeaturesByCollectionID(ctx, collectionID)
}

// AddSingleFeature добавляет новую фичу в коллекцию
func (s *GeoService) AddSingleFeature(
	ctx context.Context,
	feature *models.GeoJSONFeature,
) error {
	// проверяем существование коллекции и получаем её SRID
	col, err := s.repo.GetCollectionByID(ctx, feature.CollectionID)
	if err != nil {
		return err
	}
	if col == nil {
		return fmt.Errorf("collection %d not found", feature.CollectionID)
	}
	// вызываем репозиторий
	return s.repo.AddSingleFeature(ctx, feature, col.SRID)
}

// UpdateFeature обновляет фичу в коллекции
func (s *GeoService) UpdateFeature(ctx context.Context, feature *models.GeoJSONFeature) error {
	if feature.ID == 0 {
		return errors.New("ID фичи не установлен")
	}
	if feature.CollectionID == 0 {
		return errors.New("ID коллекции не установлен")
	}
	return s.repo.UpdateFeature(ctx, feature, srid4326)
}

func (s *GeoService) DeleteFeature(ctx context.Context, id int) error {
	feature, err := s.repo.GetFeatureByID(ctx, id)
	if err != nil {
		return err
	}
	if feature == nil {
		return errors.New("фича не найдена")
	}
	return s.repo.DeleteFeature(ctx, id)
}

func (s *GeoService) GetAllCollections(ctx context.Context) ([]*models.GeoJSONCollection, error) {
	return s.repo.GetCollections(ctx)
}

// services/geojson_service.go

func (s *GeoService) ImportGeoJSONBulk(
	ctx context.Context,
	reader io.Reader,
	name, description string,
	userID int,
) (*models.GeoJSONCollection, error) {
	// 1) создаём коллекцию как раньше
	col := &models.GeoJSONCollection{
		Name:        name,
		Description: description,
		SRID:        4326, // или другой SRID по-умолчанию
		UserID:      userID,
	}
	if err := s.repo.CreateCollection(ctx, col); err != nil {
		return nil, err
	}

	// 2) парсим весь GeoJSON из reader
	raw, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	var top struct {
		Features []struct {
			Properties json.RawMessage `json:"properties"`
			Geometry   json.RawMessage `json:"geometry"`
		} `json:"features"`
	}
	if err := json.Unmarshal(raw, &top); err != nil {
		return nil, err
	}

	// 3) готовим slice моделей
	feats := make([]*models.GeoJSONFeature, len(top.Features))
	for i, f := range top.Features {
		feats[i] = &models.GeoJSONFeature{
			Properties:   models.JSONData(f.Properties), // json.RawMessage
			Geometry:     models.JSONData(f.Geometry),   // json.RawMessage
			CollectionID: col.ID,
		}
	}

	// 4) bulk‑вставка через batch
	if err := s.repo.AddFeaturesBulk(ctx, feats, col.SRID); err != nil {
		return nil, err
	}

	return col, nil
}

// GetFeatureByID получает фичу по ID
func (s *GeoService) GetFeatureByID(ctx context.Context, id int) (*models.GeoJSONFeature, error) {
	return s.repo.GetFeatureByID(ctx, id)
}
