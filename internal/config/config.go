package config

import (
	"crypto/rsa"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
	PrivateKey  *rsa.PrivateKey
	PublicKey   *rsa.PublicKey
}

func LoadConfig() (*Config, error) {
	// Intentar cargar .env, pero no fallar si no existe
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/gama_db?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	privKeyPath := os.Getenv("JWT_PRIVATE_KEY_PATH")
	if privKeyPath == "" {
		privKeyPath = "certs/private.pem"
	}

	pubKeyPath := os.Getenv("JWT_PUBLIC_KEY_PATH")
	if pubKeyPath == "" {
		pubKeyPath = "certs/public.pem"
	}

	// Cargar llave privada RSA
	privKeyBytes, err := os.ReadFile(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("error leyendo la llave privada JWT de %s: %w", privKeyPath, err)
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("error parseando la llave privada JWT: %w", err)
	}

	// Cargar llave pública RSA
	pubKeyBytes, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return nil, fmt.Errorf("error leyendo la llave pública JWT de %s: %w", pubKeyPath, err)
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("error parseando la llave pública JWT: %w", err)
	}

	return &Config{
		DatabaseURL: dbURL,
		Port:        port,
		PrivateKey:  privKey,
		PublicKey:   pubKey,
	}, nil
}
