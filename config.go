package main

import (
	"log"
	"os"
)

type Config struct {
	DbUrl        string
	WalletUrl    string
	WalletSecret string
}

func LoadConfig() Config {
	dbURL := os.Getenv("DATABASE_URL")
	walletUrl := os.Getenv("WALLET_URL")
	walletSecret := os.Getenv("WALLET_SECRET")
	if walletUrl == "" || walletSecret == "" {
		log.Println("WALLET_URL and WALLET_SECRET must be set")
	}
	if dbURL == "" {
		dbURL = "postgres://rgs:rgs@localhost:5432/rgs?sslmode=disable"
	}

	return Config{
		DbUrl:        dbURL,
		WalletUrl:    walletUrl,
		WalletSecret: walletSecret,
	}
}
