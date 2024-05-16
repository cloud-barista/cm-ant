package spider

import (
	"fmt"

	"github.com/cloud-barista/cm-ant/pkg/configuration"
)

func SpiderHostWithPort() string {
	config := configuration.Get().Spider
	return fmt.Sprintf("%s:%s", config.Host, config.Port)
}
