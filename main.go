package main

import (
	"log"
	"net/http"
)

func main() {
	cfg := LoadConfig()
	db := ConnectDB(cfg)

	r := BuildApp(db, cfg)

	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
