package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"sync"
)

var config *AppConfig
var once sync.Once

type AppConfig struct {
	isDebug bool `yaml:"is_debug"`

	Storage struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"Port"`
		Username string `yaml:"Username"`
		Database string `yaml:"Database"`
		Password string `yaml:"Password"`
	} `yaml:"storage"`
}

func GetConfig() *AppConfig {
	return config
}

func init() {
	// This function in once run
	once.Do(func() {
		logger := logging.GetLogger()
		logger.Info("read app configuration")
		config = &AppConfig{}
		err := cleanenv.ReadConfig("config.yaml", config)
		if err != nil {
			help, _ := cleanenv.GetDescription(config, nil)
			logger.Info(help)
			logger.Fatalln(err)
		}
	})
}
