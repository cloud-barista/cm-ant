package configuration

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"strings"
	"sync"
)

var (
	appConfig *AntConfig
	once      sync.Once
)

func Get() *AntConfig {
	if appConfig == nil {
		log.Println(">>>> configuration process has not completed")

		once.Do(func() {
			Initialize()
		})
	}

	return appConfig
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
	datasource struct {
		Driver     string `yaml:"driver"`
		Connection string `yaml:"connection"`
		Username   string `yaml:"username"`
		Password   string `yaml:"password"`
	} `yaml:"datasource"`
	DB Repo
}

func Initialize() error {

	log.Println(">>>> start initialize application configuration")

	// configure app
	err := initAppConfig()
	if err != nil {
		log.Fatal(err)
		return err
	}

	// config database
	err = initDatabase()
	if err != nil {
		log.Fatal(err)
		return err
	}

	log.Println(">>>> complete initialize application configuration")

	return nil
}

func initAppConfig() error {
	log.Println(">>>> start initAppConfig()")
	cfg := AntConfig{}

	viper.AddConfigPath(RootPath())
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("fatal error config file: %w", err)
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return fmt.Errorf("fatal error config file: %w", err)
	}

	cfg.Load.JMeter.WorkDir = strings.Replace(cfg.Load.JMeter.WorkDir, "~", HomePath(), 1)
	appConfig = &cfg
	log.Println(">>>> completed initAppConfig()")

	return nil
}
