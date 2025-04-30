package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"Datapolis/internal/models"
)

const srid4326 = 4326

type GeoRepository struct {
	db *pgxpool.Pool
}

func NewGeoRepository(db *pgxpool.Pool) *GeoRepository {
	return &GeoRepository{db: db}
}

// CreateCollection создает новую коллекцию GeoJSON
func (r *GeoRepository) CreateCollection(ctx context.Context, c *models.GeoJSONCollection) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO geo_collections (name, description, srid, user_id)
         VALUES ($1,$2,$3,$4) RETURNING id, created_at, updated_at`,
		c.Name, c.Description, c.SRID, c.UserID,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *GeoRepository) GetCollections(
	ctx context.Context,
) ([]*models.GeoJSONCollection, error) {

	const q = `
	SELECT id, name, description, srid,
	       user_id, created_at, updated_at
	FROM   geo_collections
	ORDER BY created_at DESC;`

	return r.scanCollections(ctx, q)
}

// DeleteCollection удаляет коллекцию
func (r *GeoRepository) DeleteCollection(ctx context.Context, id, userID int) error {
	cmd, err := r.db.Exec(ctx, `DELETE FROM geo_collections WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("collection not found or not owned by user")
	}
	return nil
}

func (r *GeoRepository) FeaturesByCollection(ctx context.Context, collectionID int) ([]*models.GeoJSONFeature, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, properties, ST_AsGeoJSON(geometry)::jsonb AS geometry, collection_id, created_at, updated_at
           FROM geo_features WHERE collection_id=$1`, collectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*models.GeoJSONFeature
	for rows.Next() {
		f := &models.GeoJSONFeature{}
		var props, geom []byte
		if err := rows.Scan(&f.ID, &props, &geom, &f.CollectionID, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		f.Properties = models.JSONData(props)
		f.Geometry = models.JSONData(geom)
		res = append(res, f)
	}
	return res, rows.Err()
}

// DeleteFeature удаляет фичу
func (r *GeoRepository) DeleteFeature(ctx context.Context, id int) error {
	cmd, err := r.db.Exec(ctx, `DELETE FROM geo_features WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("feature not found")
	}
	return nil
}

func (r *GeoRepository) GetFeaturesByCollectionID(
	ctx context.Context,
	collectionID int,
	pagination *models.Pagination,
) ([]*models.GeoJSONFeature, error) {
	// Получаем общее количество фич
	var total int
	err := r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM geo_features WHERE collection_id = $1",
		collectionID).Scan(&total)
	if err != nil {
		return nil, err
	}
	pagination.SetTotal(total)

	rows, err := r.db.Query(ctx, `
        SELECT id,
               properties,
               ST_AsGeoJSON(geometry)::jsonb,
               collection_id,
               created_at,
               updated_at
        FROM   geo_features
        WHERE  collection_id = $1
        ORDER  BY id
        LIMIT $2 OFFSET $3`,
		collectionID, pagination.Limit, pagination.Offset())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feats []*models.GeoJSONFeature
	for rows.Next() {
		f := new(models.GeoJSONFeature)
		var props, geom []byte
		if err := rows.Scan(&f.ID, &props, &geom,
			&f.CollectionID, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		f.Properties = models.JSONData(props)
		f.Geometry = models.JSONData(geom)
		feats = append(feats, f)
	}
	return feats, rows.Err()
}

func (r *GeoRepository) GetCollectionByID(
	ctx context.Context, id int,
) (*models.GeoJSONCollection, error) {

	const q = `
	SELECT id, name, description, srid,
	       user_id, created_at, updated_at
	FROM   geo_collections
	WHERE  id = $1;`

	col := new(models.GeoJSONCollection)
	err := r.db.QueryRow(ctx, q, id).Scan(
		&col.ID,
		&col.Name,
		&col.Description,
		&col.SRID,
		&col.UserID,
		&col.CreatedAt,
		&col.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return col, nil
}

func (r *GeoRepository) UpdateFeature(
	ctx context.Context,
	f *models.GeoJSONFeature,
	srid int, // обычно defaultSRID
) error {
	props := f.Properties
	geom := f.Geometry

	res, err := r.db.Exec(ctx, `
        UPDATE geo_features
           SET properties = $1,
               geometry   = ST_SetSRID(ST_GeomFromGeoJSON($2), $3),
               updated_at = NOW()
         WHERE id = $4
           AND collection_id = $5;
    `,
		props,
		geom,
		srid,
		f.ID,
		f.CollectionID,
	)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("feature not found")
	}
	return nil
}

func (r *GeoRepository) GetFeatureByID(ctx context.Context, id int) (*models.GeoJSONFeature, error) {
	const q = `
	SELECT id,
	       collection_id,
	       properties,
	       ST_AsGeoJSON(geometry) AS geom,          -- конвертируем в GeoJSON
	       created_at,
	       updated_at
	FROM   geo_features
	WHERE  id = $1;
	`

	var (
		propsData []byte
		geoJSON   string
		f         models.GeoJSONFeature
	)

	err := r.db.QueryRow(ctx, q, id).Scan(
		&f.ID,
		&f.CollectionID,
		&propsData,
		&geoJSON,
		&f.CreatedAt,
		&f.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // не найдено — сервис вернёт «фича не найдена»
		}
		return nil, err
	}

	f.Properties = models.JSONData(propsData)
	f.Geometry = models.JSONData(geoJSON) // API‑слою отдаём уже GeoJSON

	return &f, nil
}

// helper function to scan collections
func (r *GeoRepository) scanCollections(
	ctx context.Context, query string, args ...any,
) ([]*models.GeoJSONCollection, error) {

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.GeoJSONCollection
	for rows.Next() {
		c := new(models.GeoJSONCollection)
		err = rows.Scan(
			&c.ID,
			&c.Name,
			&c.Description,
			&c.SRID,
			&c.UserID,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

// AddFeaturesBulk выполняет серию INSERT в одном батче.
func (r *GeoRepository) AddFeaturesBulk(
	ctx context.Context,
	features []*models.GeoJSONFeature,
	srid int,
) error {
	batch := &pgx.Batch{}
	for _, f := range features {
		props := json.RawMessage(`{}`)
		if len(f.Properties) > 0 {
			props = json.RawMessage(f.Properties)
		}
		batch.Queue(
			`INSERT INTO geo_features
                (properties, geometry, collection_id)
             VALUES
                ($1, ST_SetSRID(ST_GeomFromGeoJSON($2), $3), $4)
             RETURNING id, created_at, updated_at;`,
			props,
			f.Geometry,
			srid,
			f.CollectionID,
		)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for _, f := range features {
		if err := br.QueryRow().Scan(&f.ID, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return err
		}
	}

	return nil
}

func (r *GeoRepository) AddSingleFeature(
	ctx context.Context,
	f *models.GeoJSONFeature,
	srid int,
) error {
	props := f.Properties
	geom := f.Geometry

	// ST_SetSRID(ST_GeomFromGeoJSON($2), $3) конвертит GeoJSON → geometry
	err := r.db.QueryRow(ctx, `
        INSERT INTO geo_features
            (properties, geometry, collection_id)
        VALUES
            ($1,
             ST_SetSRID(ST_GeomFromGeoJSON($2), $3),
             $4)
        RETURNING id, created_at, updated_at;
    `,
		props,
		geom,
		srid,
		f.CollectionID,
	).Scan(&f.ID, &f.CreatedAt, &f.UpdatedAt)
	return err
}
