package main

import (
	"github.com/spf13/viper"
	"strings"
)

type config struct {
	Datasource struct {
		DBType         string `yaml:"dbType"`
		Url            string `yaml:"url"`
		PostgresConfig string `yaml:"postgresConfig"`
	} `yaml:"datasource"`
	Server struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		CertFile string `yaml:"certFile"`
		KeyFile  string `yaml:"keyFile"`
	} `yaml:"server"`
	Jwt struct {
		SecretKey string `yaml:"secretKey"`
	} `yaml:"jwt"`
	Logger struct {
		Profile string `yaml:"profile"`
	} `yaml:"logger"`
}

func readConfig() (*config, error) {
	// default
	viper.SetDefault("datasource.dbType", "sqlite")
	viper.SetDefault("datasource.url", "./local.db")
	viper.SetDefault("datasource.postgresConfig", "")
	viper.SetDefault("server.host", "127.0.0.1")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.certFile", "")
	viper.SetDefault("server.keyFile", "")
	viper.SetDefault("jwt.secretKey", "realworld-secret-key")
	viper.SetDefault("logger.profile", "dev")

	// yaml
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/realworld/")
	viper.AddConfigPath(".")

	// env
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	_ = viper.ReadInConfig()
	var config config
	err := viper.Unmarshal(&config)
	return &config, err
}
