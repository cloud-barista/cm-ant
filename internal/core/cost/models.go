package cost

import (
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"gorm.io/gorm"
)

type PriceInfos []*PriceInfo

type PriceInfo struct {
	gorm.Model
	ProviderName           string
	RegionName             string
	InstanceType           string
	ZoneName               string
	VCpu                   string
	Memory                 string
	MemoryUnit             constant.MemoryUnit
	OriginalMemory         string
	Storage                string
	OsType                 string
	ProductDescription     string
	OriginalPricePolicy    string
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
	MigrationId         string `gorm:"index"`
	Provider            string `gorm:"index"`
	ConnectionName      string
	ResourceType        constant.ResourceType `gorm:"index"`
	Category            string                `gorm:"index"`
	Cost                float64
	Unit                string
	ActualResourceId    string `gorm:"index"`
	FormattedResourceId string
	Granularity         string    `gorm:"index"`
	StartDate           time.Time `gorm:"index"`
	EndDate             time.Time `gorm:"index"`
}

type CostUpdateRestrict struct {
	gorm.Model
	StandardDate time.Time
	UpdateCount  int64
}
