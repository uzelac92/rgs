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
	if dbURL == "" {
		dbURL = "postgres://rgs:rgs@localhost:5432/rgs?sslmode=disable"
	}

	walletUrl := os.Getenv("WALLET_URL")
	walletSecret := os.Getenv("WALLET_SECRET")
	if walletUrl == "" || walletSecret == "" {
		log.Println("WALLET_URL and WALLET_SECRET must be set")
	}

	return Config{
		DbUrl:        dbURL,
		WalletUrl:    walletUrl,
		WalletSecret: walletSecret,
	}
}
