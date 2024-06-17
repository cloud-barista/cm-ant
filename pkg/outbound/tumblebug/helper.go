package tumblebug

import (
	"encoding/base64"
	"fmt"

	"github.com/cloud-barista/cm-ant/pkg/config"
)

func TumblebugHostWithPort() string {
	config := config.AppConfig.Tumblebug
	return fmt.Sprintf("%s:%s", config.Host, config.Port)
}

func TumblebugBaseAuthHeader() string {
	config := config.AppConfig.Tumblebug
	header := fmt.Sprintf("%s:%s", config.Username, config.Password)
	encodedHeader := base64.StdEncoding.EncodeToString([]byte(header))
	return fmt.Sprintf("Basic %s", encodedHeader)
}
