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

	FromEmail              string
	FromEmailPassword      string
	FromEmailSMTP          string
	SMTPAddress            string
	OWNER_EMAIL            string
	SUBJECT_DESC           string
	HTML_TEMPLATE          string
	OAUTH_CREDENTIALS_PATH string

	JPEG_FORMAT string
	JPG_FORMAT  string
	TXT_FORMAT  string
	JSON_FORMAT string
}

var Vars = initConfig()

func initConfig() Config {
	godotenv.Load()
	return Config{
		CLI_NAME:               GetEnv("CLI_NAME", "cryptix"),
		CLI_BINARY_PATH:        GetEnv("CLI_BINARY_PATH", "bin/cryptix"),
		CLI_VERSION:            GetEnv("CLI_VERSION", "1.0.0-stable"),
		FromEmail:              GetEnv("FROM_EMAIL", ""),
		FromEmailPassword:      GetEnv("FROM_EMAIL_PASSWORD", ""),
		FromEmailSMTP:          GetEnv("FROM_EMAIL_SMTP", "smtp.gmail.com"),
		SMTPAddress:            GetEnv("SMTP_ADDR", "smtp.gmail.com:587"),
		SUBJECT_DESC:           GetEnv("SUBJECT_DESC", "Hey smthg for you!!"),
		OAUTH_CREDENTIALS_PATH: GetEnv("CREDENTIALS_PATH", ""),
		HTML_TEMPLATE:          GetEnv("HTML_TEMPLATE", "email.html"),
		JPEG_FORMAT:            GetEnv("JPEG_FORMAT", ".jpeg"),
		JPG_FORMAT:             GetEnv("JPG_FORMAT", ".jpg"),
		TXT_FORMAT:             GetEnv("TXT_FORMAT", ".txt"),
		JSON_FORMAT:            GetEnv("JSON_FORMAT", ".json"),
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
