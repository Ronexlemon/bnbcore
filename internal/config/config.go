package config

import (
	"os"
	"strconv"
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
		DBName:                     os.Getenv("DB_NAME"),
		DBSSLMode:                  os.Getenv("DB_SSLMODE"),
		JWT_SECRET:                 os.Getenv("JWT_SECRET"),
		JWT_REFRESH_DURATION_HOUR:  refreshHour,
		JWT_ACCESS_DURATION_MINUTE: accessMin,
		MasterKeyHex: os.Getenv("MasterKeyHex"),
		EncryptedDataKeyHex: os.Getenv("EncryptedDataKeyHex"),
		PASSWORD_PEPPERS: os.Getenv("PASSWORD_PEPPERS"),
		ACTIVE_PEPPER_VERSION: os.Getenv("ACTIVE_PEPPER_VERSION"),
		GOOGLE_CLIENT_ID:os.Getenv("GOOGLE_CLIENT_ID"),
		BASE_DOMAIN:os.Getenv("BASE_DOMAIN"),

	}
}