package cost

import (
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"gorm.io/gorm"
)

type PriceInfo struct {
	gorm.Model
	ProviderName       string
	ConnectionName     string
	RegionName         string
	InstanceType       string
	ZoneName           string
	VCpu               string
	Memory             string
	MemoryUnit         constant.MemoryUnit
	OriginalMemory     string
	Storage            string
	OsType             string
	ProductDescription string

	PricePolicy            constant.PricePolicy
	Price                  string
	Currency               constant.PriceCurrency
	Unit                   constant.PriceUnit
	OriginalUnit           string
	OriginalCurrency       string
	CalculatedMonthlyPrice string
	PriceDescription       string
}

type CostInfos []CostInfo

type CostInfo struct {
	gorm.Model
	MigrationId         string
	Provider            string
	ConnectionName      string
	ResourceType        constant.ResourceType
	Category            string
	Cost                float64
	Unit                string
	ActualResourceId    string
	FormattedResourceId string
	Granularity         string
	StartDate           time.Time
	EndDate             time.Time
}

type CostUpdateRestrict struct {
	gorm.Model
	StandardDate time.Time
	UpdateCount  int64
}
