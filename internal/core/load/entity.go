package load

import "gorm.io/gorm"

type MonitoringAgentInfo struct {
	gorm.Model

	PublicIP  string
	PemKey    string
	Username  string
	Status    string
	AgentType string

	AdditionalNsId    string
	AdditionalMcisId  string
	AdditionalVmId    string
	AdditionalVmCount int
}
