package cost

import (
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"gorm.io/gorm"
)

type EstimateCostInfos []*EstimateCostInfo

type EstimateCostInfo struct {
	gorm.Model
	ProviderName           string `gorm:"index"`
	RegionName             string `gorm:"index"`
	InstanceType           string `gorm:"index"`
	ZoneName               string
	VCpu                   string `gorm:"index"`
	Memory                 string `gorm:"index"`
	MemoryUnit             constant.MemoryUnit
	OriginalMemory         string
	Storage                string
	OsType                 string `gorm:"index"`
	ProductDescription     string
	OriginalPricePolicy    string
	PricePolicy            constant.PricePolicy
	Price                  string
	Currency               constant.PriceCurrency
	Unit                   constant.PriceUnit
	OriginalUnit           string
	OriginalCurrency       string
	CalculatedMonthlyPrice float64 `gorm:"index"`
	PriceDescription       string
	LastUpdatedAt          time.Time
	ImageName              string `gorm:"index"`
}

type EstimateForecastCostInfos []EstimateForecastCostInfo

type EstimateForecastCostInfo struct {
	gorm.Model
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
	NsId                string    `gorm:"index"`
	MciId               string    `gorm:"index"`
}
