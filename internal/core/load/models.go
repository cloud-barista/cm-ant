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
	InfraId string
	NodeId  string
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
	NodeId          string
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

	// MciName is the actual cb-tumblebug MCI name backing this generator.
	// Empty means the config base name (load.defaultResourceName.mci). It is rotated to
	// base-01/-02/... by the newInstall recovery mode (FR-MA2-PERF-007-09), which leaves the
	// old MCI orphaned for the operator to clean up instead of deleting it.
	MciName string

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

	NsId    string
	InfraId string
	NodeId  string

	// NodeUid is cb-tumblebug's generated identifier for the node under test.
	//
	// NsId/InfraId/NodeId are *names*: cb-tumblebug builds a node id as "{group name}-{index
	// within the group}" and reuses it when the same names are used again. Deleting a VM and
	// recreating it with the same name therefore produces an identical NsId/InfraId/NodeId,
	// and a lookup by those alone returns the previous VM's run - a result belonging to a
	// machine that no longer exists. Only the uid is regenerated per VM, so it is what tells
	// two generations apart.
	//
	// Empty for runs recorded before this column existed; callers must treat "" as unknown
	// rather than as a mismatch.
	NodeUid string

	ExecutionStatus             constant.ExecutionStatus
	StartAt                     time.Time
	FinishAt                    *time.Time
	ExpectedFinishAt            time.Time
	TotalExpectedExcutionSecond uint64
	FailureMessage              string
	CompileDuration             string
	ExecutionDuration           string
	WithMetrics                 bool

	// not to make one to one relationship between LoadTestExecutionInfo and LoadGeneratorInstallInfo
	TestExecutionInfoId    uint
	GeneratorInstallInfoId uint

	// Steps are the per-stage progress records (FR-MA2-PERF-007-08).
	Steps []LoadTestExecutionStep `gorm:"foreignKey:LoadTestExecutionStateId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// LoadTestExecutionStep records the status of one stage of a load test run so the web
// console can show detailed, real-time progress (FR-MA2-PERF-007-08).
type LoadTestExecutionStep struct {
	gorm.Model
	LoadTestExecutionStateId uint   `gorm:"index:idx_step_state_name,unique;index"`
	LoadTestKey              string `gorm:"index"`
	Seq                      int
	Name                     constant.ExecutionStep `gorm:"index:idx_step_state_name,unique"`
	Status                   constant.StepStatus
	Attempt                  int
	StartAt                  *time.Time
	FinishAt                 *time.Time
	Message                  string // short current-status line (e.g. "Installing JMeter (retry 1)")
	Detail                   string // verbose diagnosis / error cause
}

type LoadTestExecutionInfo struct {
	gorm.Model
	LoadTestKey  string `gorm:"index:idx_info_load_test_key,unique"`
	TestName     string `gorm:"index:idx_info_load_test_name"`
	VirtualUsers string
	Duration     string
	RampUpTime   string
	RampUpSteps  string

	NsId    string
	InfraId string
	NodeId  string

	AgentHostname  string
	AgentInstalled bool

	CompileDuration            string
	ExecutionDuration          string
	LoadTestExecutionHttpInfos []LoadTestExecutionHttpInfo

	// LoadTestExecutionInfo has one LoadTestExecutionState
	LoadTestExecutionStateId uint
	LoadTestExecutionState   LoadTestExecutionState

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

type LoadTestScenarioCatalog struct {
	gorm.Model
	Name         string `gorm:"not null;index:idx_scenario_catalog_name"`
	Description  string
	VirtualUsers string `gorm:"not null"`
	Duration     string `gorm:"not null"`
	RampUpTime   string `gorm:"not null"`
	RampUpSteps  string `gorm:"not null"`
}
