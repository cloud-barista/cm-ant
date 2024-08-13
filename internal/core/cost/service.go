package cost

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/cloud-barista/cm-ant/internal/infra/outbound/spider"
	"github.com/cloud-barista/cm-ant/internal/utils"
)

type CostService struct {
	costRepo      *CostRepository
	spiderClient  *spider.SpiderClient
	costCollector CostCollector
}

func NewCostService(costRepo *CostRepository, client *spider.SpiderClient, costCollector CostCollector) *CostService {
	return &CostService{
		costRepo:      costRepo,
		spiderClient:  client,
		costCollector: costCollector,
	}
}

var onDemandPricingPolicyMap = map[string]string{
	"aws":     "OnDemand",
	"gcp":     "OnDemand",
	"azure":   "Consumption",
	"tencent": "POSTPAID_BY_HOUR",
	"alibaba": "PayAsYouGo",
	"ibm":     "quantity_tier=1",
	"ncp":     "Monthly Flat Rate",
	"ncpvpc":  "Monthly Flat Rate",
}

type GetPriceInfoParam struct {
	MigrationId    string
	ProviderName   string
	ConnectionName string
	RegionName     string
	InstanceType   string
	ZoneName       string
	VCpu           string
	Memory         string
	Storage        string
	OsType         string
	Unit           constant.PriceUnit
	Currency       constant.PriceCurrency

	TimeStandard time.Time
	PricePolicy  constant.PricePolicy
}

type AllPriceInfoResult struct {
	PriceInfoList []PriceInfoResult `json:"priceInfoList,omitempty"`
	InfoSource    string            `json:"infoSource,omitempty"`
	ResultCount   int64             `json:"resultCount"`
}

type PriceInfoResult struct {
	ID             uint   `json:"id"`
	ProviderName   string `json:"providerName"`
	ConnectionName string `json:"connectionName"`
	RegionName     string `json:"regionName"`
	InstanceType   string `json:"instanceType"`

	ZoneName           string `json:"zoneName,omitempty"`
	VCpu               string `json:"vCpu,omitempty"`
	Memory             string `json:"memory,omitempty"`
	Storage            string `json:"storage,omitempty"`
	OsType             string `json:"osType,omitempty"`
	ProductDescription string `json:"productDescription,omitempty"`

	PricePolicy            constant.PricePolicy   `json:"pricePolicy,omitempty"`
	Unit                   constant.PriceUnit     `json:"unit,omitempty"`
	Currency               constant.PriceCurrency `json:"currency,omitempty"`
	Price                  string                 `json:"price,omitempty"`
	CalculatedMonthlyPrice string                 `json:"calculatedMonthlyPrice,omitempty"`
	PriceDescription       string                 `json:"priceDescription,omitempty"`
	LastUpdatedAt          time.Time              `json:"lastUpdatedAt,omitempty"`
}

func (c *CostService) GetPriceInfo(param GetPriceInfoParam) (AllPriceInfoResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	param.TimeStandard = time.Now().AddDate(0, 0, -3).Truncate(24 * time.Hour)
	param.PricePolicy = constant.OnDemand

	var res AllPriceInfoResult

	priceInfoList, err := c.costRepo.GetAllMatchingPriceInfoList(ctx, param)
	if err != nil {
		return res, err
	}

	pil := make([]PriceInfoResult, 0)

	if len(priceInfoList) > 0 {

		for _, v := range priceInfoList {
			r := PriceInfoResult{
				ID:                     v.ID,
				ProviderName:           v.ProviderName,
				ConnectionName:         v.ConnectionName,
				RegionName:             v.RegionName,
				InstanceType:           v.InstanceType,
				ZoneName:               v.ZoneName,
				VCpu:                   v.VCpu,
				Memory:                 fmt.Sprintf("%s %s", v.Memory, v.MemoryUnit),
				Storage:                v.Storage,
				OsType:                 v.OsType,
				ProductDescription:     v.ProductDescription,
				PricePolicy:            v.PricePolicy,
				Unit:                   v.Unit,
				Currency:               v.Currency,
				Price:                  v.Price,
				CalculatedMonthlyPrice: v.CalculatedMonthlyPrice,
				PriceDescription:       v.PriceDescription,
				LastUpdatedAt:          v.UpdatedAt,
			}
			pil = append(pil, r)
		}

		res.InfoSource = "db"
		res.PriceInfoList = pil
		res.ResultCount = int64(len(pil))

		return res, nil
	}

	req := spider.PriceInfoReq{
		ConnectionName: param.ConnectionName,
		FilterList:     generateFilterList(param),
	}

	result, err := c.spiderClient.GetPriceInfoWithContext(ctx, param.RegionName, req)

	if err != nil {

		if strings.Contains(err.Error(), "you don't have any permission") {
			return res, fmt.Errorf("you don't have permission to query the price for %s", param.ProviderName)
		}
		return res, err
	}

	createdPriceInfo := make([]*PriceInfo, 0)
	if result.CloudPriceList != nil {
		for i := range result.CloudPriceList {
			p := result.CloudPriceList[i]

			if p.PriceList != nil {
				for j := range p.PriceList {

					pl := p.PriceList[j]
					productInfo := pl.ProductInfo
					vCpu := naChecker(productInfo.Vcpu)
					originalMemory := naChecker(productInfo.Memory)

					if vCpu == "" || originalMemory == "" {
						continue
					}

					memory, memoryUnit := splitMemory(originalMemory)
					zoneName := naChecker(productInfo.ZoneName)
					osType := naChecker(productInfo.OperatingSystem)
					storage := naChecker(productInfo.Storage)
					productDescription := naChecker(productInfo.Description)

					var price, originalCurrency, originalUnit, priceDescription string
					var unit constant.PriceUnit
					var currency constant.PriceCurrency

					priceInfo := pl.PriceInfo

					if priceInfo.PricingPolicies != nil {
						for k := range priceInfo.PricingPolicies {
							policy := priceInfo.PricingPolicies[k]
							convertedPrice, err := strconv.ParseFloat(policy.Price, 64)
							if err != nil {
								continue
							}

							if convertedPrice == float64(0) {
								continue
							}
							price = naChecker(policy.Price)

							if price == "" {
								continue
							}

							originalCurrency = naChecker(policy.Currency)
							originalUnit = naChecker(policy.Unit)
							unit = parseUnit(originalUnit)
							currency = parseCurrency(originalUnit)
							priceDescription = naChecker(policy.Description)

							pi := PriceInfo{
								ProviderName:           param.ProviderName,
								ConnectionName:         param.ConnectionName,
								RegionName:             param.RegionName,
								InstanceType:           param.InstanceType,
								ZoneName:               zoneName,
								VCpu:                   vCpu,
								OriginalMemory:         originalMemory,
								Memory:                 memory,
								MemoryUnit:             memoryUnit,
								Storage:                storage,
								OsType:                 osType,
								ProductDescription:     productDescription,
								PricePolicy:            constant.OnDemand,
								Price:                  price,
								Currency:               currency,
								Unit:                   unit,
								OriginalUnit:           originalUnit,
								OriginalCurrency:       originalCurrency,
								PriceDescription:       priceDescription,
								CalculatedMonthlyPrice: calculatePrice(price, unit),
							}

							createdPriceInfo = append(createdPriceInfo, &pi)
						}
					}

				}
			}
		}
	}

	if len(createdPriceInfo) > 0 {
		err := c.costRepo.InsertAllResult(ctx, param, createdPriceInfo)
		if err != nil {
			return res, nil
		}

		for _, v := range createdPriceInfo {
			r := PriceInfoResult{
				ID:                     v.ID,
				ProviderName:           v.ProviderName,
				ConnectionName:         v.ConnectionName,
				RegionName:             v.RegionName,
				InstanceType:           v.InstanceType,
				ZoneName:               v.ZoneName,
				VCpu:                   v.VCpu,
				Memory:                 fmt.Sprintf("%s %s", v.Memory, v.MemoryUnit),
				Storage:                v.Storage,
				OsType:                 v.OsType,
				ProductDescription:     v.ProductDescription,
				PricePolicy:            v.PricePolicy,
				Unit:                   v.Unit,
				Currency:               v.Currency,
				Price:                  v.Price,
				CalculatedMonthlyPrice: v.CalculatedMonthlyPrice,
				LastUpdatedAt:          v.UpdatedAt,
			}
			pil = append(pil, r)
		}

		res.InfoSource = "api"
		res.PriceInfoList = pil
		res.ResultCount = int64(len(pil))
		return res, nil
	}

	return res, nil
}

var (
	units = map[string]bool{
		"instance-hour":               true,
		"hour":                        true,
		"yrs":                         false,
		"hrs":                         true,
		"1 hour":                      true,
		"virtual processor core-hour": true,
	}
)

func parseUnit(p string) constant.PriceUnit {
	ret := constant.PerHour

	if p == "" {
		return ret
	}

	a := strings.ToLower(p)

	if v, ok := units[a]; v && ok {
		return ret
	}

	return constant.PerYear
}

func splitMemory(input string) (string, constant.MemoryUnit) {
	if input == "" {
		return "", ""
	}

	var numberPart string
	var unitPart string

	for _, char := range input {
		if unicode.IsDigit(char) || char == '.' {
			numberPart += string(char)
		} else {
			unitPart += string(char)
		}
	}

	number, err := strconv.ParseFloat(strings.TrimSpace(numberPart), 64)
	if err != nil {
		utils.LogErrorf("error while split memory : %s", err.Error())
		return "", ""
	}

	num := fmt.Sprintf("%.0f", number)
	unit := strings.ToLower(strings.TrimSpace(unitPart))

	var memoryUnit constant.MemoryUnit
	if unit == "" || unit == "gb" || unit == "gib" {
		memoryUnit = constant.GIB
	} else {
		memoryUnit = constant.GIB
	}
	return num, memoryUnit
}

func parseCurrency(p string) constant.PriceCurrency {
	if p == "" {
		return constant.USD
	}

	switch strings.ToLower(p) {
	case "usd":
		return constant.USD
	case "krw":
		return constant.KRW
	}

	return constant.USD

}

func calculatePrice(p string, unit constant.PriceUnit) string {
	if p == "" {
		return "0.000000"
	}

	value, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
	if err != nil {
		return "0.000000"
	}

	var monthlyCostStr string
	if unit == constant.PerHour {

		monthlyCost := value * 30 * 24
		monthlyCostStr = fmt.Sprintf("%.6f", monthlyCost)
	} else if unit == constant.PerYear {
		monthlyCost := value
		monthlyCostStr = fmt.Sprintf("%.6f", monthlyCost)
	}

	return monthlyCostStr
}

func generateFilterList(param GetPriceInfoParam) []spider.FilterReq {

	providerName := strings.ToLower(param.ProviderName)

	ret := []spider.FilterReq{
		{
			Key:   "instanceType",
			Value: param.InstanceType,
		},
		{
			Key:   "pricingPolicy",
			Value: onDemandPricingPolicyMap[providerName],
		},
		{
			Key:   "regionName",
			Value: param.RegionName,
		},
	}

	if strings.TrimSpace(param.ZoneName) != "" {
		ret = append(ret, spider.FilterReq{
			Key:   "zoneName",
			Value: param.ZoneName,
		})
	}

	if strings.TrimSpace(param.VCpu) != "" {
		ret = append(ret, spider.FilterReq{
			Key:   "vcpu",
			Value: param.VCpu,
		})
	}

	if strings.TrimSpace(param.Memory) != "" {
		ret = append(ret, spider.FilterReq{
			Key:   "memory",
			Value: param.Memory,
		})
	}

	if strings.TrimSpace(param.OsType) != "" {
		ret = append(ret, spider.FilterReq{
			Key:   "operatingSystem",
			Value: param.OsType,
		})
	}

	if strings.TrimSpace(param.Storage) != "" {
		ret = append(ret, spider.FilterReq{
			Key:   "storage",
			Value: param.OsType,
		})
	}

	return ret
}

func naChecker(originalValue string) string {
	trimed := strings.TrimSpace(originalValue)
	lower := strings.ToLower(trimed)

	if lower == "na" || lower == "n/a" || lower == "0" {
		return ""
	}

	return trimed
}

type UpdateCostInfoParam struct {
	MigrationId       string
	Provider          string // currently only aws
	ConnectionName    string
	StartDate         time.Time
	EndDate           time.Time
	CostResources     []CostResourceParam
	AwsAdditionalInfo AwsAdditionalInfoParam
}

type CostResourceParam struct {
	ResourceType constant.ResourceType
	ResourceIds  []string
}

type AwsAdditionalInfoParam struct {
	OwnerId string   `json:"ownerId"`
	Regions []string `json:"regions"`
}

var (
	ErrRequestResourceEmpty    = errors.New("cost request info is not enough")
	ErrCostResultEmpty         = errors.New("cost information does not exist")
	ErrCostResultFormatInvalid = errors.New("cost result does not matching with interface")
)

type UpdateCostInfoResult struct {
	FetchedDataCount int64
	UpdatedDataCount int64
}

func (c *CostService) UpdateCostInfo(param UpdateCostInfoParam) (CostInfos, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var updateCostInfoResult UpdateCostInfoResult

	r, err := c.costCollector.GetCostInfo(ctx, param)
	if err != nil {
		return nil, err
	}

	updateCostInfoResult.FetchedDataCount = int64(len(r))

	// response 저장

	return r, nil
}

// func (c *CostService) GetCostExpect(param GetPriceInfoParam) (AllPriceInfoResult, error) {
// }
