package load

import (
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"gorm.io/gorm"
)

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

type LoadGeneratorServer struct {
	gorm.Model
	Csp             string
	Region          string
	Zone            string
	PublicIp        string
	PrivateIp       string
	PublicDns       string
	MachineType     string
	Status          string
	SshPort         string
	Lat             string
	Lon             string
	Username        string
	VmId            string
	StartTime       time.Time
	AdditionalVmKey string
	Label           string

	LoadGeneratorInstallInfoID uint
	LoadGeneratorInstallInfo   LoadGeneratorInstallInfo
}

type LoadGeneratorInstallInfo struct {
	gorm.Model
	InstallLocation constant.InstallLocation
	InstallType     string
	InstallPath     string
	InstallVersion  string
	Status          string

	PublicKeyName        string
	PrivateKeyName       string
	LoadGeneratorServers []LoadGeneratorServer
}
