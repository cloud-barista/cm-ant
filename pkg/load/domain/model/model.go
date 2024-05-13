package model

import (
	"time"

	"github.com/cloud-barista/cm-ant/pkg/load/constant"
	"gorm.io/gorm"
)

type AgentInstallInfo struct {
	gorm.Model
	NsId     string
	McisId   string
	VmId     string
	Username string
	Status   string
}

type LoadEnv struct {
	gorm.Model
	InstallLocation      constant.InstallLocation      `json:"installLocation"`
	RemoteConnectionType constant.RemoteConnectionType `json:"remoteConnectionType"`
	NsId                 string                        `json:"nsId"`
	McisId               string                        `json:"mcisId"`
	VmId                 string                        `json:"vmId"`
	Username             string                        `json:"username"`
	PublicIp             string                        `json:"publicIp"`
	Cert                 string                        `json:"cert"`
}

type LoadExecutionState struct {
	gorm.Model
	LoadEnvID       uint
	LoadTestKey     string `gorm:"unique_index;not null"`
	ExecutionStatus constant.ExecutionStatus
	StartAt         time.Time
	EndAt           *time.Time
	TotalSec        uint
}

func (l *LoadExecutionState) IsFinished() bool {
	return l.ExecutionStatus != constant.Processing
}

type LoadExecutionConfig struct {
	gorm.Model
	LoadEnvID          uint
	LoadTestKey        string              `gorm:"unique_index;not null" json:"loadTestKey"`
	TestName           string              `json:"testName"`
	VirtualUsers       string              `json:"virtualUsers"`
	Duration           string              `json:"duration"`
	RampUpTime         string              `json:"rampUpTime"`
	RampUpSteps        string              `json:"rampUpSteps"`
	LoadExecutionHttps []LoadExecutionHttp `gorm:"constraint:OnUpdate:CASCADE"`
}

type LoadExecutionHttp struct {
	gorm.Model
	LoadExecutionConfigID uint
	Method                string `json:"method"`
	Protocol              string `json:"protocol"`
	Hostname              string `json:"hostname"`
	Port                  string `json:"port"`
	Path                  string `json:"path"`
	BodyData              string `json:"bodyData"`
}

type AgentInfo struct {
	gorm.Model
	Username   string `json:"username" form:"username,omitempty"`
	PublicIp   string `json:"publicIp" form:"publicIp,omitempty"`
	PemKeyPath string `json:"pemKeyPath" form:"pemKeyPath,omitempty"`
}

func NewAgentInfo() AgentInfo {
	return AgentInfo{}
}

type LoadTestStatistics struct {
	Label         string  `json:"label"`
	RequestCount  int     `json:"requestCount"`
	Average       float64 `json:"average"`
	Median        float64 `json:"median"`
	NinetyPercent float64 `json:"ninetyPercent"`
	NinetyFive    float64 `json:"ninetyFive"`
	NinetyNine    float64 `json:"ninetyNine"`
	MinTime       float64 `json:"minTime"`
	MaxTime       float64 `json:"maxTime"`
	ErrorPercent  float64 `json:"errorPercent"`
	Throughput    float64 `json:"throughput"`
	ReceivedKB    float64 `json:"receivedKB"`
	SentKB        float64 `json:"sentKB"`
}
