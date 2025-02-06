package env

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	CLI_VERSION     string
	CLI_NAME        string
	CLI_BINARY_PATH string

	FromEmail         string
	FromEmailPassword string
	FromEmailSMTP     string
	SMTPAddress       string
	OWNER_EMAIL       string
}

var Vars = initConfig()

func initConfig() Config {
	godotenv.Load()
	return Config{
		CLI_VERSION:       GetEnv("CLI_VERSION", "1.0.0-stable"),
		FromEmail:         GetEnv("FROM_EMAIL", ""),
		FromEmailPassword: GetEnv("FROM_EMAIL_PASSWORD", ""),
		FromEmailSMTP:     GetEnv("FROM_EMAIL_SMTP", "smtp.gmail.com"),
		SMTPAddress:       GetEnv("SMTP_ADDR", "smtp.gmail.com:587"),
		CLI_NAME:          GetEnv("CLI_NAME", "orcka"),
		CLI_BINARY_PATH:   GetEnv("CLI_BINARY_PATH", "bin/orcka"),
	}
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetEnvAsInt(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fallback
		}
		return i
	}
	return fallback
}
