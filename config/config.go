package config

import "github.com/spf13/viper"

var config *Config

type Config struct {
	Bot struct {
		Name     string
		Ws       string
		Token    string
		Email    string
		Password string
	}
}

func GetConfig() (*Config, error) {

	if config != nil {
		return config, nil
	}

	viper.SetConfigName("config.toml")
	viper.SetConfigType("toml")
	viper.AddConfigPath("./_local")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/")
	viper.AddConfigPath("$HOME/.config/")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	c := Config{}
	err = viper.Unmarshal(&c)
	if err != nil {
		return nil, err
	}
	config = &c
	return &c, nil
}
