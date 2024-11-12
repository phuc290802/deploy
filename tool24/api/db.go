package main

import (
	"context"
	"fmt"
	"os"

	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
)

func NewDb() (*Db, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, fmt.Errorf("Error loading .env file: %v", err)
	}

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=verify-full", user, password, host, port, dbname)
	pool, err := pgxpool.Connect(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to the database: %v", err)
	}

	return &Db{Conn: pool}, nil
}
