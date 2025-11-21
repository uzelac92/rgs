package main

import (
	"log"
	"net/http"
	"rgs/sqlc"
)

func main() {
	cfg := LoadConfig()
	db := ConnectDB(cfg)

	q := sqlc.New(db)

	r := SetupRouter(q)

	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
