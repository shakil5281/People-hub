package config

import (
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
	return &Config{
		Port:      getEnv("API_PORT", "5000"),
		DBHost:    getEnv("DB_HOST", "localhost"),
		DBPort:    getEnv("DB_PORT", "5432"),
		DBUser:    getEnv("DB_USER", "shakil"),
		DBPass:    getEnv("DB_PASS", "123456"),
		DBName:    getEnv("DB_NAME", "peoplehub"),
		DBSSLMode: getEnv("DB_SSLMODE", "disable"),
		JWTSecret: getEnv("JWT_SECRET", "peoplehub-secret-key-change-in-production-2025"),
	}
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
