package model

import (
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"gorm.io/gorm"
)

type LoadEnv struct {
	gorm.Model
	InstallLocation constant.InstallLocation `json:"installLocation"`
	NsId            string                   `json:"nsId"`
	McisId          string                   `json:"mcisId"`
	VmId            string                   `json:"vmId"`
	Username        string                   `json:"username"`
	PublicIp        string                   `json:"publicIp"`
	PemKeyPath      string                   `json:"pemKeyPath"`
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
