package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/cloud-barista/cm-ant/internal/utils"
	"github.com/rs/zerolog/log"
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

	Cost struct {
		Estimation struct {
			UpdateInterval time.Duration `yaml:"updateInterval"`
		} `yaml:"estimation"`
	} `yaml:"cost"`
	Load struct {
		Retry   int `yaml:"retry"`
		Timeout struct {
			MonitoringAgentInstall string `yaml:"monitoringAgentInstall"`
			CommandExecution       string `yaml:"commandExecution"`
			UninstallAgent         string `yaml:"uninstallAgent"`
		} `yaml:"timeout"`
		DefaultResourceName struct {
			Namespace string `yaml:"namespace"`
			Mci       string `yaml:"mci"`
			Vm        string `yaml:"vm"`
			SshKey    string `yaml:"sshKey"`
		} `yaml:"defaultResourceName"`
		Generator struct {
			// Recovery controls what happens when a reused load generator turns out to be
			// unusable (VM deleted externally, or unreachable after retries):
			//   "auto"       - force-reset the stale generator and recreate it (default)
			//   "manual"     - stop with test_failed and a diagnostic message; operator fixes it
			//   "newInstall" - leave the old MCI orphaned, install a fresh one under a rotated
			//                  name (base-01/-02/...); the suffix is persisted in the DB.
			// FR-MA2-PERF-007-09.
			Recovery string `yaml:"recovery"`
			// Idle controls what happens to the shared load generator after a run finishes:
			//   "keep"      - leave it running for fast reuse (default; previous behavior)
			//   "suspend"   - suspend the VM; the next run resumes it
			//   "terminate" - tear the generator down; the next run recreates it (~15 min)
			// Applied only after the final result rsync completes, remote generators only.
			// FR-MA2-PERF-007-01 (BAR-1413).
			Idle string `yaml:"idle"`
		} `yaml:"generator"`
		Limits struct {
			// Upper bounds for load test parameters (FR-MA2-PERF-007-01). A value of 0 (unset)
			// falls back to the built-in default. Override via config.yaml or ANT_LOAD_LIMITS_*.
			MaxVirtualUsers int `yaml:"maxVirtualUsers"`
			MaxDuration     int `yaml:"maxDuration"`
			MaxRampUpTime   int `yaml:"maxRampUpTime"`
			MaxRampUpSteps  int `yaml:"maxRampUpSteps"`
		} `yaml:"limits"`
		JMeter struct {
			Dir     string `yaml:"dir"`
			Version string `yaml:"version"`
		} `yaml:"jmeter"`
		Image struct {
			UseSmartMatching      bool                         `yaml:"useSmartMatching"`
			UseFallbackImagesOnly bool                         `yaml:"useFallbackImagesOnly"`
			PreferredOs           string                       `yaml:"preferredOs"`
			FallbackOs            string                       `yaml:"fallbackOs"`
			SearchOptions         ImageSearchOptions           `yaml:"searchOptions"`
			OsKeywords            map[string][]string          `yaml:"osKeywords"`
			FallbackImages        map[string]map[string]string `yaml:"fallbackImages"`
		} `yaml:"image"`
		Spec struct {
			MinVcpu   int `yaml:"minVcpu"`
			MaxVcpu   int `yaml:"maxVcpu"`
			MinMemory int `yaml:"minMemory"`
			MaxMemory int `yaml:"maxMemory"`
			// Provider 제거 - 동적으로 기존 VM의 CSP 정보에서 추출
			Architecture string `yaml:"architecture"`
		} `yaml:"spec"`
	} `yaml:"load"`
	Log struct {
		Level string `yaml:"level"`
	} `yaml:"log"`
	Database struct {
		Driver   string `yaml:"driver"`
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
	} `yaml:"database"`
}

// ImageSearchOptions defines image search configuration options
type ImageSearchOptions struct {
	IsRegisteredByAsset    bool `yaml:"isRegisteredByAsset"`
	IncludeDeprecatedImage bool `yaml:"includeDeprecatedImage"`
	IncludeBasicImageOnly  bool `yaml:"includeBasicImageOnly"`
	MaxResults             int  `yaml:"maxResults"`
}

func InitConfig() error {
	log.Info().Msg("Initializing configuration...")

	cfg := AntConfig{}

	viper.AddConfigPath(utils.RootPath())
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.SetEnvPrefix("ant")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Error().Msgf("Fatal error while reading config file: %v", err)
		return fmt.Errorf("fatal error while read config file: %w", err)
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Error().Msgf("Fatal error while unmarshaling config: %v", err)
		return fmt.Errorf("fatal error while unmarshal from config to ant config: %w", err)
	}

	log.Info().Msgf("Configuration loaded successfully: %+v", cfg)
	AppConfig = cfg

	return nil
}
