package bootstrap

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Environment struct {
	DataBase   DataBase      `mapstructure:",squash"`
	Redis      Redis         `mapstructure:",squash"`
	HttpServer HttpServer    `mapstructure:",squash"`
	JWT        JWTConfig     `mapstructure:",squash"`
	OTP        OTPConfig     `mapstructure:",squash"`
	Seed       SeedConfig    `mapstructure:",squash"`
	Captcha    CaptchaConfig `mapstructure:",squash"`
}

type CaptchaConfig struct {
	TTL time.Duration `mapstructure:"CAPTCHA_TTL"`
}

type SeedConfig struct {
	UserName     string `mapstructure:"SEED_USER_NAME"`
	UserEmail    string `mapstructure:"SEED_USER_EMAIL"`
	UserPhone    string `mapstructure:"SEED_USER_PHONE"`
	UserPassword string `mapstructure:"SEED_USER_PASSWORD"`

	AdminName     string `mapstructure:"SEED_ADMIN_NAME"`
	AdminEmail    string `mapstructure:"SEED_ADMIN_EMAIL"`
	AdminPhone    string `mapstructure:"SEED_ADMIN_PHONE"`
	AdminPassword string `mapstructure:"SEED_ADMIN_PASSWORD"`
}

type JWTConfig struct {
	SecretKey       string        `mapstructure:"JWT_SECRET_KEY"`
	AccessTokenTTL  time.Duration `mapstructure:"ACCESS_TOKEN_TTL"`
	RefreshTokenTTL time.Duration `mapstructure:"REFRESH_TOKEN_TTL"`
}

type OTPConfig struct {
	TTL         time.Duration `mapstructure:"OTP_TTL"`
	VerifiedTTL time.Duration `mapstructure:"OTP_VERIFIED_TTL"`
	Digits      int           `mapstructure:"OTP_DIGITS"`
}

type DataBase struct {
	Host     string `mapstructure:"PDB_HOST"`
	Port     string `mapstructure:"PDB_PORT"`
	User     string `mapstructure:"PDB_USER"`
	Password string `mapstructure:"PDB_PASSWORD" json:"-"`
	Name     string `mapstructure:"PDB_DB_NAME"`
	SSLMode  string `mapstructure:"PDB_SSL_MODE"`
}

type Redis struct {
	Host      string `mapstructure:"RDB_HOST"`
	Port      string `mapstructure:"RDB_PORT"`
	Password  string `mapstructure:"RDB_PASSWORD" json:"-"`
	RDBNumber string `mapstructure:"RDB_DB_NUMBER"`
}

type HttpServer struct {
	Domain string `mapstructure:"HTTP_DOMAIN"`
	Port   string `mapstructure:"HTTP_PORT"`
}

func NewEnvironment() (*Environment, error) {
	v := viper.New()
	v.SetConfigFile(".env")
	v.SetConfigType("env")

	for _, key := range EnvKeys {
		if err := v.BindEnv(key); err != nil {
			return nil, err
		}
	}
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read .env: %w", err)
	}
	var environment Environment
	if err := v.Unmarshal(&environment); err != nil {
		return nil, fmt.Errorf("read environment settings: %w", err)
	}

	return &environment, nil
}
