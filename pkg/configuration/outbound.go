package configuration

import (
	"encoding/base64"
	"fmt"
)

func TumblebugHostWithPort() string {
	config := Get().Tumblebug
	return fmt.Sprintf("%s:%s", config.Host, config.Port)
}

func TumblebugBaseAuthHeader() string {
	config := Get().Tumblebug
	header := fmt.Sprintf("%s:%s", config.Username, config.Password)
	encodedHeader := base64.StdEncoding.EncodeToString([]byte(header))
	return fmt.Sprintf("Basic %s", encodedHeader)
}
