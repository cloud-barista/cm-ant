package load

import "gorm.io/gorm"

type MonitoringAgentInfo struct {
	gorm.Model

	Username  string
	Status    string
	AgentType string

	NsId    string
	McisId  string
	VmId    string
	VmCount int
}
