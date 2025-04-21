package db

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"log"
	"os"
)

var Pool *pgxpool.Pool

func ConnectDB() {
	var err error

	_ = godotenv.Load(".env")

	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		log.Println("DATABASE_URL не установлен")
		return
	}

	config, err := pgxpool.ParseConfig(databaseUrl)
	if err != nil {
		log.Printf("Ошибка при разборе URL базы данных: %v\n", err)
		return
	}

	config.ConnConfig.TLSConfig = &tls.Config{
		InsecureSkipVerify: true, // В производственной среде лучше true заменить на false
	}

	Pool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Printf("Невозможно подключиться к базе данных: %v\n", err)
		return
	}

	log.Println("Успешно подключено к базе данных")

	// Проверяем соединение
	if err := Pool.Ping(context.Background()); err != nil {
		log.Printf("Не удалось проверить соединение: %v\n", err)
		return
	}

	// Создаем таблицы
	if err := createTables(); err != nil {
		log.Printf("Ошибка при создании таблиц: %v\n", err)
	}
}

func createTables() error {
	_, err := Pool.Exec(context.Background(), `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        username VARCHAR(255) UNIQUE NOT NULL,
        password VARCHAR(255) NOT NULL,
        email VARCHAR(255) UNIQUE NOT NULL,
        role VARCHAR(50) DEFAULT 'user',
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    `)

	_, err = Pool.Exec(context.Background(), `
    CREATE TABLE IF NOT EXISTS geo_json_collections (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		type VARCHAR(50) DEFAULT 'FeatureCollection',
		crs JSONB,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
        );
    `)

	_, err = Pool.Exec(context.Background(), `
    CREATE TABLE IF NOT EXISTS geo_json_features (
		id SERIAL PRIMARY KEY,
		type VARCHAR(50) DEFAULT 'Feature',
		properties JSONB,
		geometry JSONB NOT NULL,
		collection_id INTEGER NOT NULL REFERENCES geo_json_collections(id) ON DELETE CASCADE,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
        );
    `)

	if err != nil {
		return fmt.Errorf("не удалось создать таблицы: %w", err)
	}
	log.Println("Таблицы успешно созданы")
	return nil
}
