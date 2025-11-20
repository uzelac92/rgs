package main

import (
	"log"
	"net/http"
)

func main() {
	cfg := LoadConfig()

	_ = ConnectDB(cfg)
	log.Println("Connected to PostgreSQL")

	r := SetupRouter()

	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
