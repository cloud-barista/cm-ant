package configuration

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"strings"
	"sync"
)

var (
	runtimeConfig *AntConfig
	once          sync.Once
)

func Get() *AntConfig {
	if runtimeConfig == nil {
		log.Println("configuration process has not completed")

		once.Do(func() {
			InitConfig("")
		})
	}

	return runtimeConfig
}

type AntConfig struct {
	Spider struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"spider"`
	Tumblebug struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"tumblebug"`
	Load struct {
		JMeter struct {
			WorkDir string `yaml:"workDir"`
			Version string `yaml:"version"`
		} `yaml:"jmeter"`
	} `yaml:"load"`
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
	cfg := AntConfig{}

	if configPath == "" {
		configPath = RootPath()
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

	cfg.Load.JMeter.WorkDir = strings.Replace(cfg.Load.JMeter.WorkDir, "~", HomePath(), 1)

	runtimeConfig = &cfg
	return nil
}
