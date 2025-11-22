package main

import (
	"context"
	"log"
	"net/http"
	"rgs/observability"

	"go.uber.org/zap"
)

func main() {
	if err := observability.InitLogger(); err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer func(Logger *zap.Logger) {
		err := Logger.Sync()
		if err != nil {
			log.Fatalf("failed to sync logger: %v", err)
		}
	}(observability.Logger)

	shutdownTracing := observability.InitTracing()
	defer shutdownTracing(context.Background())

	observability.InitMetrics()

	cfg := LoadConfig()
	db := ConnectDB(cfg)

	r := BuildApp(db, cfg)

	observability.Logger.Info("Server starting", zap.String("port", "8080"))
	if err := http.ListenAndServe(":8080", r); err != nil {
		observability.Logger.Fatal("Error starting the server", zap.Error(err))
	}
}
