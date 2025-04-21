package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"Datapolis/internal/models"
)

type GeoJSONRepository struct {
	db *pgxpool.Pool
}

func NewGeoJSONRepository(db *pgxpool.Pool) *GeoJSONRepository {
	return &GeoJSONRepository{db: db}
}

// CreateCollection создает новую коллекцию GeoJSON
func (r *GeoJSONRepository) CreateCollection(ctx context.Context, collection *models.GeoJSONCollection) error {
	crsData, err := json.Marshal(collection.CRS)
	if err != nil {
		return err
	}

	err = r.db.QueryRow(ctx,
		`INSERT INTO geo_json_collections (name, description, type, crs, user_id) 
		VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`,
		collection.Name, collection.Description, collection.Type, crsData, collection.UserID).
		Scan(&collection.ID, &collection.CreatedAt, &collection.UpdatedAt)
	return err
}

// GetCollectionByID получает коллекцию по ID
func (r *GeoJSONRepository) GetCollectionByID(ctx context.Context, id int) (*models.GeoJSONCollection, error) {
	collection := &models.GeoJSONCollection{}
	var crsData []byte

	err := r.db.QueryRow(ctx,
		`SELECT id, name, description, type, crs, user_id, created_at, updated_at 
		FROM geo_json_collections WHERE id = $1`, id).
		Scan(&collection.ID, &collection.Name, &collection.Description, &collection.Type,
			&crsData, &collection.UserID, &collection.CreatedAt, &collection.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	collection.CRS = models.JSONData(crsData)
	return collection, nil
}

// GetCollectionsByUserID получает все коллекции пользователя
func (r *GeoJSONRepository) GetCollectionsByUserID(ctx context.Context, userID int) ([]*models.GeoJSONCollection, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, description, type, crs, user_id, created_at, updated_at 
		FROM geo_json_collections WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collections []*models.GeoJSONCollection
	for rows.Next() {
		collection := &models.GeoJSONCollection{}
		var crsData []byte

		err := rows.Scan(&collection.ID, &collection.Name, &collection.Description, &collection.Type,
			&crsData, &collection.UserID, &collection.CreatedAt, &collection.UpdatedAt)
		if err != nil {
			return nil, err
		}

		collection.CRS = models.JSONData(crsData)
		collections = append(collections, collection)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return collections, nil
}

// UpdateCollection обновляет коллекцию
func (r *GeoJSONRepository) UpdateCollection(ctx context.Context, collection *models.GeoJSONCollection) error {
	crsData, err := json.Marshal(collection.CRS)
	if err != nil {
		return err
	}

	result, err := r.db.Exec(ctx,
		`UPDATE geo_json_collections SET name = $1, description = $2, type = $3, crs = $4, updated_at = NOW() 
		WHERE id = $5 AND user_id = $6`,
		collection.Name, collection.Description, collection.Type, crsData, collection.ID, collection.UserID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("коллекция не найдена или нет прав на редактирование")
	}

	return nil
}

// DeleteCollection удаляет коллекцию
func (r *GeoJSONRepository) DeleteCollection(ctx context.Context, id int, userID int) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM geo_json_collections WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("коллекция не найдена или нет прав на удаление")
	}

	return nil
}

// AddFeature добавляет новую фичу в коллекцию
func (r *GeoJSONRepository) AddFeature(ctx context.Context, feature *models.GeoJSONFeature) error {
	propertiesData, err := json.Marshal(feature.Properties)
	if err != nil {
		return err
	}

	geometryData, err := json.Marshal(feature.Geometry)
	if err != nil {
		return err
	}

	err = r.db.QueryRow(ctx,
		`INSERT INTO geo_json_features (type, properties, geometry, collection_id) 
		VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`,
		feature.Type, propertiesData, geometryData, feature.CollectionID).
		Scan(&feature.ID, &feature.CreatedAt, &feature.UpdatedAt)
	return err
}

// GetFeaturesByCollectionID получает все фичи коллекции
func (r *GeoJSONRepository) GetFeaturesByCollectionID(ctx context.Context, collectionID int) ([]*models.GeoJSONFeature, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, type, properties, geometry, collection_id, created_at, updated_at 
		FROM geo_json_features WHERE collection_id = $1`, collectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var features []*models.GeoJSONFeature
	for rows.Next() {
		feature := &models.GeoJSONFeature{}
		var propertiesData, geometryData []byte

		err := rows.Scan(&feature.ID, &feature.Type, &propertiesData, &geometryData,
			&feature.CollectionID, &feature.CreatedAt, &feature.UpdatedAt)
		if err != nil {
			return nil, err
		}

		feature.Properties = models.JSONData(propertiesData)
		feature.Geometry = models.JSONData(geometryData)
		features = append(features, feature)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return features, nil
}

// UpdateFeature обновляет фичу
func (r *GeoJSONRepository) UpdateFeature(ctx context.Context, feature *models.GeoJSONFeature) error {
	propertiesData, err := json.Marshal(feature.Properties)
	if err != nil {
		return err
	}

	geometryData, err := json.Marshal(feature.Geometry)
	if err != nil {
		return err
	}

	result, err := r.db.Exec(ctx,
		`UPDATE geo_json_features SET type = $1, properties = $2, geometry = $3, updated_at = NOW() 
		WHERE id = $4 AND collection_id = $5`,
		feature.Type, propertiesData, geometryData, feature.ID, feature.CollectionID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("фича не найдена")
	}

	return nil
}

// DeleteFeature удаляет фичу
func (r *GeoJSONRepository) DeleteFeature(ctx context.Context, id int, collectionID int) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM geo_json_features WHERE id = $1 AND collection_id = $2`, id, collectionID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("фича не найдена")
	}

	return nil
}

func (r *GeoJSONRepository) GetFeatureByID(ctx context.Context, id int) (*models.GeoJSONFeature, error) {
	feature := &models.GeoJSONFeature{}
	var propertiesData, geometryData []byte

	err := r.db.QueryRow(ctx,
		`SELECT id, type, properties, geometry, collection_id, created_at, updated_at
  FROM geo_json_features WHERE id = $1`, id).
		Scan(&feature.ID, &feature.Type, &propertiesData, &geometryData,
			&feature.CollectionID, &feature.CreatedAt, &feature.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	feature.Properties = models.JSONData(propertiesData)
	feature.Geometry = models.JSONData(geometryData)
	return feature, nil
}
