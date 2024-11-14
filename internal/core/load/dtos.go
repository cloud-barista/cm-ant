package load

import (
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
)

// MonitoringAgentInstallationParams represents parameters for installing a monitoring agent.
type MonitoringAgentInstallationParams struct {
	NsId  string   `json:"nsId"`
	MciId string   `json:"mciId"`
	VmIds []string `json:"vmIds,omitempty"`
}

// MonitoringAgentInstallationResult represents the result of a monitoring agent installation.
type MonitoringAgentInstallationResult struct {
	ID        uint      `json:"id,omitempty"`
	NsId      string    `json:"nsId,omitempty"`
	MciId     string    `json:"mciId,omitempty"`
	VmId      string    `json:"vmId,omitempty"`
	VmCount   int       `json:"vmCount,omitempty"`
	Status    string    `json:"status,omitempty"`
	Username  string    `json:"username,omitempty"`
	AgentType string    `json:"agentType,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

type GetAllMonitoringAgentInfosParam struct {
	Page  int    `json:"page"`
	Size  int    `json:"size"`
	NsId  string `json:"nsId,omitempty"`
	MciId string `json:"mciId,omitempty"`
	VmId  string `json:"vmId,omitempty"`
}

type GetAllMonitoringAgentInfoResult struct {
	MonitoringAgentInfos []MonitoringAgentInstallationResult `json:"monitoringAgentInfos,omitempty"`
	TotalRow             int64                               `json:"totalRow,omitempty"`
}

type InstallLoadGeneratorParam struct {
	InstallLocation constant.InstallLocation `json:"installLocation,omitempty"`
	Coordinates     []string                 `json:"coordinate"`
}

type LoadGeneratorServerResult struct {
	ID              uint      `json:"id,omitempty"`
	Csp             string    `json:"csp,omitempty"`
	Region          string    `json:"region,omitempty"`
	Zone            string    `json:"zone,omitempty"`
	PublicIp        string    `json:"publicIp,omitempty"`
	PrivateIp       string    `json:"privateIp,omitempty"`
	PublicDns       string    `json:"publicDns,omitempty"`
	MachineType     string    `json:"machineType,omitempty"`
	Status          string    `json:"status,omitempty"`
	SshPort         string    `json:"sshPort,omitempty"`
	Lat             string    `json:"lat,omitempty"`
	Lon             string    `json:"lon,omitempty"`
	Username        string    `json:"username,omitempty"`
	VmId            string    `json:"vmId,omitempty"`
	StartTime       string    `json:"startTime,omitempty"`
	AdditionalVmKey string    `json:"additionalVmKey,omitempty"`
	Label           string    `json:"label,omitempty"`
	CreatedAt       time.Time `json:"createdAt,omitempty"`
	UpdatedAt       time.Time `json:"updatedAt,omitempty"`
}

type LoadGeneratorInstallInfoResult struct {
	ID              uint                     `json:"id,omitempty"`
	InstallLocation constant.InstallLocation `json:"installLocation,omitempty"`
	InstallType     string                   `json:"installType,omitempty"`
	InstallPath     string                   `json:"installPath,omitempty"`
	InstallVersion  string                   `json:"installVersion,omitempty"`
	Status          string                   `json:"status,omitempty"`
	CreatedAt       time.Time                `json:"createdAt,omitempty"`
	UpdatedAt       time.Time                `json:"updatedAt,omitempty"`

	PublicKeyName        string                      `json:"publicKeyName,omitempty"`
	PrivateKeyName       string                      `json:"privateKeyName,omitempty"`
	LoadGeneratorServers []LoadGeneratorServerResult `json:"loadGeneratorServers,omitempty"`
}

type UninstallLoadGeneratorParam struct {
	LoadGeneratorInstallInfoId uint
}

type GetAllLoadGeneratorInstallInfoParam struct {
	Page   int    `json:"page"`
	Size   int    `json:"size"`
	Status string `json:"Status"`
}

type GetAllLoadGeneratorInstallInfoResult struct {
	LoadGeneratorInstallInfoResults []LoadGeneratorInstallInfoResult `json:"loadGeneratorInstallInfoResults,omitempty"`
	TotalRows                       int64                            `json:"totalRows,omitempty"`
}

type RunLoadTestParam struct {
	LoadTestKey                string                    `json:"loadTestKey"`
	InstallLoadGenerator       InstallLoadGeneratorParam `json:"installLoadGenerator"`
	LoadGeneratorInstallInfoId uint                      `json:"loadGeneratorInstallInfoId"`

	// test scenario
	TestName     string `json:"testName"`
	VirtualUsers string `json:"virtualUsers"`
	Duration     string `json:"duration"`
	RampUpTime   string `json:"rampUpTime"`
	RampUpSteps  string `json:"rampUpSteps"`

	// related tumblebug
	NsId  string `json:"nsId"`
	MciId string `json:"mciId"`
	VmId  string `json:"vmId"`

	CollectAdditionalSystemMetrics bool
	AgentHostname                  string

	HttpReqs []RunLoadTestHttpParam `json:"httpReqs,omitempty"`
}

type RunLoadTestHttpParam struct {
	Method   string `json:"method"`
	Protocol string `json:"protocol"`
	Hostname string `json:"hostname"`
	Port     string `json:"port"`
	Path     string `json:"path,omitempty"`
	BodyData string `json:"bodyData,omitempty"`
}

type GetAllLoadTestExecutionStateParam struct {
	Page            int                      `json:"page"`
	Size            int                      `json:"size"`
	LoadTestKey     string                   `json:"loadTestKey"`
	ExecutionStatus constant.ExecutionStatus `json:"executionStatus"`
}

type GetAllLoadTestExecutionStateResult struct {
	LoadTestExecutionStates []LoadTestExecutionStateResult `json:"loadTestExecutionStates,omitempty"`
	TotalRow                int64                          `json:"totalRow,omitempty"`
}

type LoadTestExecutionStateResult struct {
	ID                          uint                           `json:"id"`
	LoadGeneratorInstallInfoId  uint                           `json:"loadGeneratorInstallInfoId,omitempty"`
	LoadGeneratorInstallInfo    LoadGeneratorInstallInfoResult `json:"loadGeneratorInstallInfo,omitempty"`
	LoadTestKey                 string                         `json:"loadTestKey,omitempty"`
	ExecutionStatus             constant.ExecutionStatus       `json:"executionStatus,omitempty"`
	StartAt                     time.Time                      `json:"startAt,omitempty"`
	FinishAt                    *time.Time                     `json:"finishAt,omitempty"`
	ExpectedFinishAt            time.Time                      `json:"expectedFinishAt,omitempty"`
	IconCode                    constant.IconCode              `json:"iconCode"`
	TotalExpectedExcutionSecond uint64                         `json:"totalExpectedExecutionSecond,omitempty"`
	FailureMessage              string                         `json:"failureMessage,omitempty"`
	CompileDuration             string                         `json:"compileDuration,omitempty"`
	ExecutionDuration           string                         `json:"executionDuration,omitempty"`
	CreatedAt                   time.Time                      `json:"createdAt,omitempty"`
	UpdatedAt                   time.Time                      `json:"updatedAt,omitempty"`
}

type GetLoadTestExecutionStateParam struct {
	LoadTestKey string `json:"loadTestKey"`
	NsId        string `json:"nsId"`
	MciId       string `json:"mciId"`
	VmId        string `json:"vmId"`
}

type GetAllLoadTestExecutionInfosParam struct {
	Page int `json:"page"`
	Size int `json:"size"`
}

type GetAllLoadTestExecutionInfosResult struct {
	TotalRow               int64                         `json:"totalRow,omitempty"`
	LoadTestExecutionInfos []LoadTestExecutionInfoResult `json:"loadTestExecutionInfos,omitempty"`
}

type LoadTestExecutionInfoResult struct {
	ID                         uint                              `json:"id"`
	LoadTestKey                string                            `json:"loadTestKey,omitempty"`
	TestName                   string                            `json:"testName,omitempty"`
	VirtualUsers               string                            `json:"virtualUsers,omitempty"`
	Duration                   string                            `json:"duration,omitempty"`
	RampUpTime                 string                            `json:"rampUpTime,omitempty"`
	RampUpSteps                string                            `json:"rampUpSteps,omitempty"`
	AgentHostname              string                            `json:"agentHostname,omitempty"`
	AgentInstalled             bool                              `json:"agentInstalled,omitempty"`
	CompileDuration            string                            `json:"compileDuration,omitempty"`
	ExecutionDuration          string                            `json:"executionDuration,omitempty"`
	LoadTestExecutionHttpInfos []LoadTestExecutionHttpInfoResult `json:"loadTestExecutionHttpInfos,omitempty"`
	LoadTestExecutionState     LoadTestExecutionStateResult      `json:"loadTestExecutionState,omitempty"`
	LoadGeneratorInstallInfo   LoadGeneratorInstallInfoResult    `json:"loadGeneratorInstallInfo,omitempty"`
}

type LoadTestExecutionHttpInfoResult struct {
	ID       uint   `json:"id"`
	Method   string `json:"method,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	Port     string `json:"port,omitempty"`
	Path     string `json:"path,omitempty"`
	BodyData string `json:"bodyData,omitempty"`
}

type GetLoadTestExecutionInfoParam struct {
	LoadTestKey string `json:"loadTestKey"`
}

type StopLoadTestParam struct {
	LoadTestKey string `json:"loadTestKey"`
	NsId        string `json:"nsId"`
	MciId       string `json:"mciId"`
	VmId        string `json:"vmId"`
}

type ResultSummary struct {
	Label   string
	Results []*ResultRawData
}

type MetricsSummary struct {
	Label   string
	Metrics []*MetricsRawData
}

type ResultRawData struct {
	No         int
	Elapsed    int // time to last byte
	Bytes      int
	SentBytes  int
	URL        string
	Latency    int // time to first byte
	IdleTime   int // time not spent sampling in jmeter (milliseconds) (generally 0)
	Connection int // time to establish connection
	IsError    bool
	Timestamp  time.Time
}

type MetricsRawData struct {
	Value     string
	Unit      string
	IsError   bool
	Timestamp time.Time
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

type GetLoadTestResultParam struct {
	LoadTestKey string
	Format      constant.ResultFormat
}

type GetLastLoadTestResultParam struct {
	NsId   string
	MciId  string
	VmId   string
	Format constant.ResultFormat
}
