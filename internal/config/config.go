package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
	DBSSLMode string
	JWTSecret string
}

func init() {
	loadEnvFiles()
}

func loadEnvFiles() {
	root := findProjectRoot()
	if root == "" {
		return
	}

	loadIfExists := func(name string) {
		path := filepath.Join(root, name)
		if _, err := os.Stat(path); err != nil {
			return
		}
		envMap, err := godotenv.Read(path)
		if err != nil {
			return
		}
		for key, val := range envMap {
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}

	loadIfExists(".env")

	goEnv := os.Getenv("GO_ENV")
	if goEnv == "production" {
		loadIfExists(".env.production")
	} else {
		loadIfExists(".env.local")
	}
}

func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func Load() *Config {
	cfg := &Config{
		Port:      getEnv("API_PORT", "5000"),
		DBHost:    getEnv("DB_HOST", "localhost"),
		DBPort:    getEnv("DB_PORT", "5432"),
		DBUser:    getEnv("DB_USER", "shakil"),
		DBPass:    getEnv("DB_PASS", "123456"),
		DBName:    getEnv("DB_NAME", "peoplehub"),
		DBSSLMode: getEnv("DB_SSLMODE", "disable"),
		JWTSecret: getEnv("JWT_SECRET", "peoplehub-secret-key-change-in-production-2025"),
	}
	// Production must fail fast on insecure defaults - never silently use hardcoded secrets
	if os.Getenv("GO_ENV") == "production" {
		if cfg.JWTSecret == "" || cfg.JWTSecret == "peoplehub-secret-key-change-in-production-2025" || cfg.JWTSecret == "CHANGEME_min32chars_replace_with_openssl_rand_hex_32" || len(cfg.JWTSecret) < 32 {
			log.Fatal("FATAL: JWT_SECRET must be set to a strong value (>=32 chars) in production. Generate with: openssl rand -hex 32")
		}
		if cfg.DBPass == "" || cfg.DBPass == "123456" || cfg.DBPass == "CHANGEME_STRONG_DB_PASSWORD_min16chars" || len(cfg.DBPass) < 12 {
			log.Fatal("FATAL: DB_PASS must be set to a strong value in production")
		}
		if cfg.DBSSLMode == "disable" {
			log.Println("WARN: DB_SSLMODE=disable in production - consider enable with proper certs")
		}
	} else {
		// Development warning for insecure defaults
		if cfg.JWTSecret == "peoplehub-secret-key-change-in-production-2025" {
			log.Println("WARN: Using default JWT_SECRET - set a strong JWT_SECRET in .env.local for development")
		}
		if cfg.DBPass == "123456" {
			log.Println("WARN: Using default DB_PASS=123456 - set a strong password in .env.local")
		}
	}
	return cfg
}

func (c *Config) GetDSN() string {
	return "host=" + c.DBHost +
		" port=" + c.DBPort +
		" user=" + c.DBUser +
		" password=" + c.DBPass +
		" dbname=" + c.DBName +
		" sslmode=" + c.DBSSLMode
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
