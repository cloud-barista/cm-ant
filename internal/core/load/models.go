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

	IsCluster   bool
	IsMaster    bool
	ClusterSize uint64

	LoadGeneratorInstallInfoId uint
	LoadGeneratorInstallInfo   LoadGeneratorInstallInfo
}

type LoadGeneratorInstallInfo struct {
	gorm.Model
	InstallLocation constant.InstallLocation
	InstallType     string
	InstallPath     string
	InstallVersion  string
	Status          string

	IsCluster   bool
	MasterId    uint
	ClusterSize uint64

	PublicKeyName        string
	PrivateKeyName       string
	LoadGeneratorServers []LoadGeneratorServer
}

type LoadTestExecutionState struct {
	gorm.Model
	LoadTestKey                 string `gorm:"unique_index;not null"`
	ExecutionStatus             constant.ExecutionStatus
	StartAt                     time.Time
	FinishAt                    *time.Time
	TotalExpectedExcutionSecond uint64
	FailureMessage              string
	CompileDuration             string `json:"compileDuration"`
	ExecutionDuration           string `json:"executionDuration"`

	LoadTestExecutionInfoId uint

	LoadGeneratorInstallInfoId uint
	LoadGeneratorInstallInfo   LoadGeneratorInstallInfo
}

type LoadTestExecutionInfo struct {
	gorm.Model
	LoadTestKey                string                      `json:"loadTestKey" gorm:"unique_index;not null"`
	TestName                   string                      `json:"testName"`
	VirtualUsers               string                      `json:"virtualUsers"`
	Duration                   string                      `json:"duration"`
	RampUpTime                 string                      `json:"rampUpTime"`
	RampUpSteps                string                      `json:"rampUpSteps"`
	Hostname                   string                      `json:"hostname"`
	Port                       string                      `json:"port"`
	AgentHostname              string                      `json:"agentHostname"`
	AgentInstalled             bool                        `json:"agentInstalled"`
	CompileDuration            string                      `json:"compileDuration"`
	ExecutionDuration          string                      `json:"executionDuration"`
	LoadTestExecutionHttpInfos []LoadTestExecutionHttpInfo `json:"httpReqs,omitempty"`

	LoadTestExecutionState LoadTestExecutionState

	LoadGeneratorInstallInfoId uint
	LoadGeneratorInstallInfo   LoadGeneratorInstallInfo
}

type LoadTestExecutionHttpInfo struct {
	gorm.Model
	Method   string `json:"method"`
	Protocol string `json:"protocol"`
	Hostname string `json:"hostname"`
	Port     string `json:"port"`
	Path     string `json:"path,omitempty"`
	BodyData string `json:"bodyData,omitempty"`

	LoadTestExecutionInfoId uint
}
