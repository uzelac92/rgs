package main

import (
	"database/sql"
	"rgs/observability"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func ConnectDB(cfg Config) *sql.DB {
	db, err := sql.Open("pgx", cfg.DbUrl)
	if err != nil {
		observability.Logger.Fatal("failed to open db", zap.Error(err))
	}

	if err := db.Ping(); err != nil {
		observability.Logger.Fatal("failed to ping db", zap.Error(err))
	}

	return db
}
