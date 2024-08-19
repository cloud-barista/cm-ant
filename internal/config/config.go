package config

import (
	"fmt"
	"strings"

	"github.com/cloud-barista/cm-ant/internal/utils"
	"github.com/spf13/viper"
)

var (
	AppConfig AntConfig
)

type AntConfig struct {
	Root struct {
		Path string `yaml:"path"`
	} `yaml:"root"`
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Spider struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"spider"`
	Tumblebug struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"tumblebug"`
	Load struct {
		Retry  int `yaml:"retry"`
		JMeter struct {
			Dir     string `yaml:"dir"`
			Version string `yaml:"version"`
		} `yaml:"jmeter"`
	} `yaml:"load"`
	Logging struct {
		Level string `yaml:"level"`
	} `yaml:"logging"`
	Database struct {
		Driver   string `yaml:"driver"`
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
	} `yaml:"database"`
}

func InitConfig() error {
	utils.LogInfo("Initializing configuration...")

	cfg := AntConfig{}

	viper.AddConfigPath(utils.RootPath())
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.SetEnvPrefix("ant")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		utils.LogErrorf("Fatal error while reading config file: %v", err)
		return fmt.Errorf("fatal error while read config file: %w", err)
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		utils.LogErrorf("Fatal error while unmarshaling config: %v", err)
		return fmt.Errorf("fatal error while unmarshal from config to ant config: %w", err)
	}

	utils.LogInfof("Configuration loaded successfully: %+v", cfg)
	AppConfig = cfg

	return nil
}
