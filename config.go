package main

import (
	"os"
)

type Config struct {
	DbUrl string
}

func LoadConfig() Config {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://rgs:rgs@localhost:5432/rgs?sslmode=disable"
	}

	return Config{
		DbUrl: dbURL,
	}
}
