package bootstrap

import (
	"fmt"

	"github.com/spf13/viper"
)

type Environment struct {
	DataBase   DataBase   `mapstructure:",squash"`
	Redis      Redis      `mapstructure:",squash"`
	HttpServer HttpServer `mapstructure:",squash"`
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
	for _, key := range []string{
		"HTTP_DOMAIN", "HTTP_PORT",
		"PDB_HOST", "PDB_PORT", "PDB_USER", "PDB_PASSWORD", "PDB_DB_NAME", "PDB_SSL_MODE",
		"RDB_HOST", "RDB_PORT", "RDB_PASSWORD", "RDB_DB_NUMBER",
	} {
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
