package cost

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/cloud-barista/cm-ant/internal/infra/outbound/spider"
	"github.com/cloud-barista/cm-ant/internal/utils"
)

type PriceCollector interface {
	GetPriceInfos(context.Context, GetPriceInfoParam) (PriceInfos, error)
}

var (
	onDemandPricingPolicyMap = map[string]string{
		"aws":     "OnDemand",
		"gcp":     "OnDemand",
		"azure":   "Consumption",
		"tencent": "POSTPAID_BY_HOUR",
		"alibaba": "PayAsYouGo",
		"ibm":     "quantity_tier=1",
		"ncp":     "Monthly Flat Rate",
		"ncpvpc":  "Monthly Flat Rate",
	}

	priceValidator = map[string]func(res *PriceInfo) bool{
		"aws": func(res *PriceInfo) bool {
			return !strings.Contains(res.PriceDescription, "Reservation")
		},
		"gcp":   func(res *PriceInfo) bool { return true },
		"azure": func(res *PriceInfo) bool { return true },
		"tencent": func(res *PriceInfo) bool {
			return res.OriginalPricePolicy == onDemandPricingPolicyMap[res.ProviderName]
		},
		"alibaba": func(res *PriceInfo) bool { return true },
		"ibm": func(res *PriceInfo) bool {
			return strings.Contains(res.OriginalUnit, "Instance-Hour")
		},
		"ncp":    func(res *PriceInfo) bool { return true },
		"ncpvpc": func(res *PriceInfo) bool { return true },
	}

	units = map[string]bool{
		"instance-hour":               true,
		"hour":                        true,
		"yrs":                         false,
		"hrs":                         true,
		"1 hour":                      true,
		"virtual processor core-hour": true,
	}
)

type SpiderPriceCollector struct {
	sc *spider.SpiderClient
}

func NewSpiderPriceCollector(sc *spider.SpiderClient) PriceCollector {
	return &SpiderPriceCollector{
		sc: sc,
	}
}

func (s *SpiderPriceCollector) GetPriceInfos(ctx context.Context, param GetPriceInfoParam) (PriceInfos, error) {
	connectionName := fmt.Sprintf("%s-%s", strings.ToLower(param.ProviderName), strings.ToLower(param.RegionName))

	req := spider.PriceInfoReq{
		ConnectionName: connectionName,
		FilterList:     s.generateFilterList(param),
	}

	result, err := s.sc.GetPriceInfoWithContext(ctx, param.RegionName, req)

	if err != nil {

		if strings.Contains(err.Error(), "you don't have any permission") {
			return nil, fmt.Errorf("you don't have permission to query the price for %s", param.ProviderName)
		}
		return nil, err
	}

	createdPriceInfo := make([]*PriceInfo, 0)
	if result.CloudPriceList != nil {
		for i := range result.CloudPriceList {
			p := result.CloudPriceList[i]

			if p.PriceList != nil {
				for j := range p.PriceList {

					pl := p.PriceList[j]

					productInfo := pl.ProductInfo
					vCpu := s.naChecker(productInfo.Vcpu)
					originalMemory := s.naChecker(productInfo.Memory)

					if vCpu == "" || originalMemory == "" {
						continue
					}

					memory, memoryUnit := s.splitMemory(originalMemory)
					zoneName := s.naChecker(productInfo.ZoneName)
					osType := s.naChecker(productInfo.OperatingSystem)
					storage := s.naChecker(productInfo.Storage)
					productDescription := s.naChecker(productInfo.Description)

					var price, originalCurrency, originalUnit, priceDescription string
					var unit constant.PriceUnit
					var currency constant.PriceCurrency

					priceInfo := pl.PriceInfo

					if priceInfo.PricingPolicies != nil {
						for k := range priceInfo.PricingPolicies {
							policy := priceInfo.PricingPolicies[k]
							originalPricePolicy := s.naChecker(policy.PricingPolicy)
							priceDescription = s.naChecker(policy.Description)
							originalCurrency = s.naChecker(policy.Currency)
							originalUnit = s.naChecker(policy.Unit)
							unit = s.parseUnit(originalUnit)
							currency = s.parseCurrency(policy.Currency)
							convertedPrice, err := strconv.ParseFloat(policy.Price, 64)
							if err != nil {
								continue
							}

							if convertedPrice == float64(0) {
								continue
							}
							price = s.naChecker(policy.Price)

							if price == "" {
								continue
							}

							pi := PriceInfo{
								ProviderName:           param.ProviderName,
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
								OriginalPricePolicy:    originalPricePolicy,
								PricePolicy:            constant.OnDemand,
								Price:                  price,
								Currency:               currency,
								Unit:                   unit,
								OriginalUnit:           originalUnit,
								OriginalCurrency:       originalCurrency,
								PriceDescription:       priceDescription,
								CalculatedMonthlyPrice: s.calculatePrice(price, unit),
							}

							if !priceValidator[param.ProviderName](&pi) {
								continue
							}

							createdPriceInfo = append(createdPriceInfo, &pi)
						}
					}

				}
			}
		}
	}

	sort.Slice(createdPriceInfo, func(i, j int) bool {
		return createdPriceInfo[i].Price < createdPriceInfo[j].Price
	})

	return createdPriceInfo, nil
}

func (s *SpiderPriceCollector) generateFilterList(param GetPriceInfoParam) []spider.FilterReq {

	providerName := strings.ToLower(param.ProviderName)
	param.ProviderName = providerName

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

	return ret
}

func (s *SpiderPriceCollector) parseUnit(p string) constant.PriceUnit {
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

func (s *SpiderPriceCollector) splitMemory(input string) (string, constant.MemoryUnit) {
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

func (s *SpiderPriceCollector) parseCurrency(p string) constant.PriceCurrency {
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

func (s *SpiderPriceCollector) calculatePrice(p string, unit constant.PriceUnit) string {
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

func (s *SpiderPriceCollector) naChecker(originalValue string) string {
	trimed := strings.TrimSpace(originalValue)
	lower := strings.ToLower(trimed)

	if lower == "na" || lower == "n/a" || lower == "0" {
		return ""
	}

	return trimed
}
