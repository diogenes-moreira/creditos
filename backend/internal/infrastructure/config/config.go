package config

import (
	"os"
	"strconv"
)

type Config struct {
	DB     DBConfig
	Server ServerConfig
	JWT    JWTConfig
	GCS    GCSConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	Instance string
}

func (c DBConfig) DSN() string {
	sslMode := c.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	dsn := "host=" + c.Host +
		" user=" + c.User +
		" password=" + c.Password +
		" dbname=" + c.Name +
		" port=" + c.Port +
		" sslmode=" + sslMode +
		" TimeZone=America/Argentina/Buenos_Aires"
	return dsn
}

type ServerConfig struct {
	Port string
	Mode string
}

type JWTConfig struct {
	Secret          string
	ExpirationHours int
}

type GCSConfig struct {
	Bucket          string
	CredentialsFile string
	LocalPath       string
}

func Load() *Config {
	return &Config{
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "creditos"),
			Password: getEnv("DB_PASSWORD", "creditos_secret"),
			Name:     getEnv("DB_NAME", "creditos"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			Instance: getEnv("DB_INSTANCE", ""),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "change-me-in-production"),
			ExpirationHours: getEnvInt("JWT_EXPIRATION_HOURS", 24),
		},
		GCS: GCSConfig{
			Bucket:          getEnv("GCS_BUCKET", ""),
			CredentialsFile: getEnv("GCS_CREDENTIALS_FILE", ""),
			LocalPath:       getEnv("LOCAL_STORAGE_PATH", "./storage"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
