package configuration

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	runtimeConfig *AntConfig
	_, b, _, _    = runtime.Caller(0)
	basePath      = filepath.Dir(b)
	lock          sync.RWMutex
)

func Get() *AntConfig {
	if runtimeConfig == nil {
		log.Println("configuration process has not completed")

		lock.Lock()
		defer lock.Unlock()
		InitConfig("")
	}

	return runtimeConfig
}

type AntConfig struct {
	Spider struct {
		URL  string `yaml:"url"`
		Port string `yaml:"port"`
	} `yaml:"spider"`
	Tumblebug struct {
		URL      string `yaml:"url"`
		Port     string `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"tumblebug"`
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Datasource struct {
		Type     string `yaml:"type"`
		URL      string `yaml:"url"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"datasource"`
}

func InitConfig(configPath string) error {
	if runtimeConfig != nil {
		return errors.New("cm-ant is already configured")
	}

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
