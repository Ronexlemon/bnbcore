package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DBHost                     string
	DBPort                     string
	DBUser                     string
	DBPassword                 string
	DBName                     string
	DBSSLMode                  string
	JWT_SECRET                 string
	JWT_REFRESH_DURATION_HOUR  int 
	JWT_ACCESS_DURATION_MINUTE int 
	MasterKeyHex string
    EncryptedDataKeyHex string
	PASSWORD_PEPPERS string
	ACTIVE_PEPPER_VERSION string
	GOOGLE_CLIENT_ID string
	BASE_DOMAIN string
	TWILIO_ACCOUNT_SID string
TWILIO_AUTH_TOKEN string
TWILIO_WHATSAPP_FROM  string
HOST string
PORT string   
USERNAME string
PASSWORD string
FROM string
REDIS_URL string
CLOUDINARY_URL string
TEMPLATE_SID string
REDIS_PASSWORD string
BASE_APP_URL string

}

func Load() *Config {

	accessMin, err := strconv.Atoi(os.Getenv("JWT_ACCESS_DURATION_MINUTE"))
	if err != nil {
		accessMin = 15 
	}

	
	refreshHour, err := strconv.Atoi(os.Getenv("JWT_REFRESH_DURATION_HOUR"))
	if err != nil {
		refreshHour = 7 
	}

	return &Config{
		DBHost:                     os.Getenv("DB_HOST"),
		DBPort:                     os.Getenv("DB_PORT"),
		DBUser:                     os.Getenv("DB_USER"),
		DBPassword:                 os.Getenv("DB_PASSWORD"),
		DBName:                    strings.TrimSpace(os.Getenv("DB_NAME")),
		DBSSLMode:                  strings.TrimSpace(os.Getenv("DB_SSLMODE")),
		JWT_SECRET:                 strings.TrimSpace(os.Getenv("JWT_SECRET")),
		JWT_REFRESH_DURATION_HOUR:  refreshHour,
		JWT_ACCESS_DURATION_MINUTE: accessMin,
		MasterKeyHex: strings.TrimSpace(os.Getenv("MasterKeyHex")),
		EncryptedDataKeyHex: strings.TrimSpace(os.Getenv("EncryptedDataKeyHex")),
		PASSWORD_PEPPERS: strings.TrimSpace(os.Getenv("PASSWORD_PEPPERS")),
		ACTIVE_PEPPER_VERSION:strings.TrimSpace(os.Getenv("ACTIVE_PEPPER_VERSION")),
		GOOGLE_CLIENT_ID:strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID")),
		BASE_DOMAIN:strings.TrimSpace(os.Getenv("BASE_DOMAIN")),
		TWILIO_ACCOUNT_SID:strings.TrimSpace(os.Getenv("TWILIO_ACCOUNT_SID")),
		TWILIO_AUTH_TOKEN:strings.TrimSpace(os.Getenv("TWILIO_AUTH_TOKEN")),
		TWILIO_WHATSAPP_FROM:strings.TrimSpace(os.Getenv("TWILIO_WHATSAPP_FROM")),
		HOST:strings.TrimSpace(os.Getenv("HOST")),
		PORT:strings.TrimSpace(os.Getenv("PORT")),
		FROM:strings.TrimSpace(os.Getenv("FROM")),
		PASSWORD:strings.TrimSpace(os.Getenv("PASSWORD")),
		USERNAME:strings.TrimSpace(os.Getenv("USERNAME")),
		REDIS_URL:strings.TrimSpace(os.Getenv("REDIS_URL")),
		CLOUDINARY_URL:strings.TrimSpace(os.Getenv("CLOUDINARY_URL")),
		TEMPLATE_SID :strings.TrimSpace(os.Getenv("TEMPLATE_SID")),
		REDIS_PASSWORD :strings.TrimSpace(os.Getenv("REDIS_PASSWORD")),
		BASE_APP_URL :strings.TrimSpace(os.Getenv("BASE_APP_URL")),
		

	}
}