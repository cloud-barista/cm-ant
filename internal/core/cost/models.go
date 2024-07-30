package cost

import (
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
