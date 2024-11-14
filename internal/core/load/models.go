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
	MciId   string
	VmId    string
	VmCount int
}

type LoadGeneratorServer struct {
	gorm.Model
	VmUid           string
	VmName          string
	ImageName       string
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
	StartTime       string
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
	LoadGeneratorServers []LoadGeneratorServer `gorm:"foreignKey:LoadGeneratorInstallInfoId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type LoadTestExecutionState struct {
	gorm.Model
	LoadTestKey string `gorm:"index:idx_state_load_test_key,unique"`

	NsId  string
	MciId string
	VmId  string

	ExecutionStatus             constant.ExecutionStatus
	StartAt                     time.Time
	FinishAt                    *time.Time
	ExpectedFinishAt            time.Time
	TotalExpectedExcutionSecond uint64
	FailureMessage              string
	CompileDuration             string
	ExecutionDuration           string

	// not to make one to one relationship between LoadTestExecutionInfo and LoadGeneratorInstallInfo
	TestExecutionInfoId uint
	GeneratorInstallInfoId uint
}

type LoadTestExecutionInfo struct {
	gorm.Model
	LoadTestKey  string `gorm:"index:idx_info_load_test_key,unique"`
	TestName     string `gorm:"index:idx_info_load_test_name"`
	VirtualUsers string
	Duration     string
	RampUpTime   string
	RampUpSteps  string

	NsId  string
	MciId string
	VmId  string

	AgentHostname  string
	AgentInstalled bool

	CompileDuration            string
	ExecutionDuration          string
	LoadTestExecutionHttpInfos []LoadTestExecutionHttpInfo

	// LoadTestExecutionInfo has one LoadTestExecutionState
	LoadTestExecutionStateId uint
	LoadTestExecutionState LoadTestExecutionState

	// LoadTestExecutionInfo has one LoadGeneratorInstallInfo
	LoadGeneratorInstallInfoId uint
	LoadGeneratorInstallInfo   LoadGeneratorInstallInfo
}

type LoadTestExecutionHttpInfo struct {
	gorm.Model
	Method   string
	Protocol string
	Hostname string
	Port     string
	Path     string
	BodyData string

	LoadTestExecutionInfoId uint
}
