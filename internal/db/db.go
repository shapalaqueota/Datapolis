package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"log"
	"os"
)

var Pool *pgxpool.Pool

func ConnectDB() {

	var err error
	godotenv.Load(".env")
	databaseUrl := os.Getenv("DATABASE_URL")

	config, err := pgxpool.ParseConfig(databaseUrl)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to parse database URL: %v \n", err))
	}

	Pool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to connect to database: %v \n", err))
	}

	log.Println("Successfully connected to database")
	createTables()
}

func createTables() {
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
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to create tables: %v \n", err))
	}
	log.Println("Successfully created tables")
}
