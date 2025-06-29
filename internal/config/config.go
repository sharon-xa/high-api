package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Env struct {
	// App
	Environment           string `mapstructure:"APP_ENV"`
	Port                  int    `mapstructure:"PORT"`
	AdminEmail            string `mapstructure:"ADMIN_EMAIL"`
	OtpExpMin             int    `mapstructure:"OTP_EXP_IN_MIN"`
	PasswordResetExpInMin int    `mapstructure:"PASSWORD_RESET_EXP_IN_MIN"`
	APIDomain             string `mapstructure:"API_DOMAIN"`
	UserMinAge            int    `mapstructure:"USER_MIN_AGE"`

	// Email
	Email       string `mapstructure:"EMAIL"`
	Password    string `mapstructure:"PASSWORD"`
	FrontendURL string `mapstructure:"FRONTEND_PASS_RESET_URL"`

	// DB
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBName     string `mapstructure:"DB_DATABASE"`
	DBUsername string `mapstructure:"DB_USERNAME"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBSchema   string `mapstructure:"DB_SCHEMA"`
	DBTimeZone string `mapstructure:"DB_TIMEZONE"`
	DSN        string

	// S3
	S3AccessKey       string `mapstructure:"S3_ACCESS_KEY_ID"`
	S3SecretAccessKey string `mapstructure:"S3_SECRET_ACCESS_KEY"`
	S3Region          string `mapstructure:"S3_REGION"`
	S3Bucket          string `mapstructure:"S3_BUCKET"`

	// JWT
	TokenSecret           string `mapstructure:"TOKEN_SECRET"`
	AccessTokenSecret     string `mapstructure:"ACCESS_TOKEN_SECRET"`
	RefreshTokenSecret    string `mapstructure:"REFRESH_TOKEN_SECRET"`
	AccessTokenExpInMin   int    `mapstructure:"ACCESS_TOKEN_EXP_IN_MIN"`
	RefreshTokenExpInDays int    `mapstructure:"REFRESH_TOKEN_EXP_IN_DAYS"`

	// Hash
	HashSecret string `mapstructure:"HASHING_SECRET"`
}

func NewEnv() *Env {
	env := Env{}
	viper.AddConfigPath("./")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Can't find the file .env:", err)
	}

	err = viper.Unmarshal(&env)
	if err != nil {
		log.Fatal("Environment can't be loaded:", err)
	}

	env.DSN = fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s",
		env.DBHost,
		env.DBUsername,
		env.DBPassword,
		env.DBName,
		env.DBPort,
		env.DBTimeZone,
	)

	if env.Environment == "dev" {
		log.Println("The App is running in development mode")
	}

	return &env
}
