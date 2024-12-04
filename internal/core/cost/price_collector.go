package cost

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/cloud-barista/cm-ant/internal/infra/outbound/spider"
	"github.com/rs/zerolog/log"
)

type PriceCollector interface {
	Readyz(context.Context) error
	FetchPriceInfos(context.Context, RecommendSpecParam) (EstimateCostInfos, error)
}

const (
	productFamily    = "ComputeInstance"
	ncpProductFamily = "Server"
)

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

	priceValidator = map[string]func(res *EstimateCostInfo) bool{
		"aws": func(res *EstimateCostInfo) bool {
			return !strings.Contains(res.PriceDescription, "Reservation")
		},
		"gcp":   func(res *EstimateCostInfo) bool { return true },
		"azure": func(res *EstimateCostInfo) bool { return true },
		"tencent": func(res *EstimateCostInfo) bool {
			return res.OriginalPricePolicy == onDemandPricingPolicyMap[res.ProviderName]
		},
		"alibaba": func(res *EstimateCostInfo) bool { return true },
		"ibm": func(res *EstimateCostInfo) bool {
			return strings.Contains(res.OriginalUnit, "Instance-Hour")
		},
		"ncp":    func(res *EstimateCostInfo) bool { return true },
		"ncpvpc": func(res *EstimateCostInfo) bool { return true },
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

func (s *SpiderPriceCollector) Readyz(ctx context.Context) error {
	err := s.sc.ReadyzWithContext(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *SpiderPriceCollector) FetchPriceInfos(ctx context.Context, param RecommendSpecParam) (EstimateCostInfos, error) {
	connectionName := fmt.Sprintf("%s-%s", strings.ToLower(param.ProviderName), strings.ToLower(param.RegionName))

	req := spider.PriceInfoReq{
		ConnectionName: connectionName,
		FilterList:     s.generateFilter(param),
	}
	pf := productFamily

	if strings.Contains(strings.ToLower(param.ProviderName), "ncp") {
		pf = ncpProductFamily
	}

	result, err := s.sc.GetPriceInfoWithContext(ctx, pf, param.RegionName, req)

	if err != nil {

		if strings.Contains(err.Error(), "you don't have any permission") {
			return nil, fmt.Errorf("you don't have permission to query the price for %s", param.ProviderName)
		}
		return nil, err
	}

	createdPriceInfo := make([]*EstimateCostInfo, 0)
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
								log.Warn().Msgf("not allowed for error; %s", err)
								continue
							}

							if convertedPrice == float64(0) {
								log.Warn().Msg("not allowed for empty price")
								continue
							}
							price = s.naChecker(policy.Price)

							if price == "" {
								log.Warn().Msg("not allowed for empty price")
								continue
							}

							if strings.Contains(strings.ToLower(priceDescription), "dedicated") {
								log.Warn().Msgf("not allowed for dedicated instance hour; %s", priceDescription)
								continue
							}

							pi := EstimateCostInfo{
								ProviderName:           param.ProviderName,
								RegionName:             productInfo.RegionName,
								InstanceType:           productInfo.InstanceType,
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
								LastUpdatedAt:          time.Now(),
								ImageName:              param.Image,
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

func (s *SpiderPriceCollector) generateFilter(param RecommendSpecParam) []spider.FilterReq {

	providerName := strings.ToLower(param.ProviderName)
	param.ProviderName = providerName

	ret := []spider.FilterReq{
		{
			Key:   "pricingPolicy",
			Value: onDemandPricingPolicyMap[providerName],
		},
		{
			Key:   "regionName",
			Value: param.RegionName,
		},
		{
			Key:   "instanceType",
			Value: param.InstanceType,
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
		log.Error().Msgf("error while split memory : %s", err.Error())
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

func (s *SpiderPriceCollector) calculatePrice(p string, unit constant.PriceUnit) float64 {
	if p == "" {
		return 0.000000
	}

	value, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
	if err != nil {
		return 0.000000
	}

	var monthlyCost float64
	if unit == constant.PerHour {
		monthlyCost = value * 30 * 24
	} else if unit == constant.PerYear {
		monthlyCost = value
	}

	roundedCost := math.Round(monthlyCost*1e6) / 1e6

	return roundedCost
}

func (s *SpiderPriceCollector) naChecker(originalValue string) string {
	trimed := strings.TrimSpace(originalValue)
	lower := strings.ToLower(trimed)

	if lower == "na" || lower == "n/a" || lower == "0" {
		return ""
	}

	return trimed
}
