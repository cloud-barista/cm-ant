package configuration

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"path/filepath"
	"runtime"
)

var (
	runtimeConfig *AntConfig
	_, b, _, _    = runtime.Caller(0)
	basePath      = filepath.Dir(b)
)

func Get() (*AntConfig, error) {
	if runtimeConfig == nil {
		log.Println("configuration process has not completed")
		return nil, errors.New("configuration process has not completed")
	}

	return runtimeConfig, nil
}

type AntConfig struct {
	Spider struct {
		URL  string `yaml:"url"`
		Port int    `yaml:"port"`
	} `yaml:"spider"`
	Tumblebug struct {
		URL  string `yaml:"url"`
		Port int    `yaml:"port"`
	} `yaml:"tumblebug"`
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Datasource struct {
		Type     string `yaml:"type"`
		URL      string `yaml:"url"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"datasource"`
}

func InitConfig(configPath string) error {
	cfg := AntConfig{}

	if configPath == "" {
		configPath = basePath[0 : len(basePath)-len("/internal/common/configuration")]
	}

	viper.AddConfigPath(configPath)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
		return fmt.Errorf("fatal error config file: %w", err)
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatal(err)
		return fmt.Errorf("fatal error config file: %w", err)
	}

	runtimeConfig = &cfg
	log.Printf("configuration completed; %v", cfg)

	return nil
}
