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

	createdPriceInfo := s.buildEstimateCostInfos(param, result)

	log.Info().Msgf("CB-Spider price result for %s %s %s: received %d items (legacy cloud price lists %d), added %d price infos",
		param.ProviderName, param.RegionName, param.InstanceType, len(result.PriceList), len(result.CloudPriceList), len(createdPriceInfo))

	return createdPriceInfo, nil
}

// buildEstimateCostInfos turns a cb-spider price response into estimate cost infos.
// It performs no I/O, so it can be tested with recorded responses.
func (s *SpiderPriceCollector) buildEstimateCostInfos(param RecommendSpecParam, result spider.CloudPriceDataRes) EstimateCostInfos {
	createdPriceInfo := make([]*EstimateCostInfo, 0)

	// cb-spider v0.11.5+ response structure (PriceList at the top level)
	for j := range result.PriceList {
		createdPriceInfo = append(createdPriceInfo, s.priceInfosFromItem(param, result.PriceList[j])...)
	}

	// cb-spider v0.10.0 response structure (CloudPriceList)
	for i := range result.CloudPriceList {
		p := result.CloudPriceList[i]

		for j := range p.PriceList {
			productInfo := p.PriceList[j].ProductInfo
			vCpu := s.naChecker(productInfo.Vcpu)
			originalMemory := s.naChecker(productInfo.Memory)

			if vCpu == "" || originalMemory == "" {
				log.Debug().Msgf("Skipping due to empty vCpu or Memory: vCpu=%s, Memory=%s", vCpu, originalMemory)
				continue
			}

			memory, memoryUnit := s.splitMemory(originalMemory)
			base := s.newEstimateCostInfo(param, productInfo, productInfo.InstanceType, vCpu, originalMemory, memory, memoryUnit)

			createdPriceInfo = append(createdPriceInfo, s.pricingPolicyPrices(param, base, p.PriceList[j].PriceInfo.PricingPolicies)...)
		}
	}

	sort.Slice(createdPriceInfo, func(i, j int) bool {
		return createdPriceInfo[i].Price < createdPriceInfo[j].Price
	})

	return createdPriceInfo
}

// priceInfosFromItem extracts the prices of one v0.11.5+ price item that matches the requested instance type.
// Sources are tried in order, and a later source is used only when the earlier one yielded no price:
//  1. CSPPriceInfo.terms.OnDemand (AWS only; other providers put a different structure in CSPPriceInfo)
//  2. PriceInfo.OnDemand (GCP and other non-AWS providers)
//  3. PriceInfo.PricingPolicies (cb-spider < v0.11.5)
func (s *SpiderPriceCollector) priceInfosFromItem(param RecommendSpecParam, pl spider.PriceRes) []*EstimateCostInfo {
	productInfo := pl.ProductInfo

	var vCpu, originalMemory, instanceType, memory string
	var memoryUnit constant.MemoryUnit

	if productInfo.VMSpecInfo != nil {
		vCpu = s.naChecker(productInfo.VMSpecInfo.VCpu.Count)
		originalMemory = s.naChecker(productInfo.VMSpecInfo.MemSizeMiB)
		instanceType = s.naChecker(productInfo.VMSpecInfo.Name)
		memory, memoryUnit = s.mibToGiB(originalMemory)
	} else {
		vCpu = s.naChecker(productInfo.Vcpu)
		originalMemory = s.naChecker(productInfo.Memory)
		instanceType = s.naChecker(productInfo.InstanceType)
		memory, memoryUnit = s.splitMemory(originalMemory)
	}

	if !strings.EqualFold(instanceType, param.InstanceType) {
		return nil
	}

	log.Debug().Msgf("Match found: got=%s, requested=%s, vCpu=%s, originalMemory=%s", instanceType, param.InstanceType, vCpu, originalMemory)

	if vCpu == "" || originalMemory == "" || memory == "" {
		log.Debug().Msgf("Skipping due to empty vCpu or Memory: vCpu=%s, Memory=%s", vCpu, originalMemory)
		return nil
	}

	base := s.newEstimateCostInfo(param, productInfo, instanceType, vCpu, originalMemory, memory, memoryUnit)
	priceInfo := pl.PriceInfo

	if param.ProviderName == "aws" {
		if ret, found := s.awsTermsPrices(param, base, priceInfo.CSPPriceInfo); found {
			return ret
		}
	}

	if pi := s.onDemandPrice(param, base, priceInfo.OnDemand); pi != nil {
		return []*EstimateCostInfo{pi}
	}

	return s.pricingPolicyPrices(param, base, priceInfo.PricingPolicies)
}

// newEstimateCostInfo fills the product fields shared by every price of one item.
func (s *SpiderPriceCollector) newEstimateCostInfo(param RecommendSpecParam, productInfo spider.ProductInfoRes, instanceType, vCpu, originalMemory, memory string, memoryUnit constant.MemoryUnit) EstimateCostInfo {
	return EstimateCostInfo{
		ProviderName:       param.ProviderName,
		RegionName:         resolveRegion(productInfo.VMSpecInfo, param.RegionName),
		InstanceType:       instanceType,
		ZoneName:           s.naChecker(productInfo.ZoneName),
		VCpu:               vCpu,
		OriginalMemory:     originalMemory,
		Memory:             memory,
		MemoryUnit:         memoryUnit,
		Storage:            s.naChecker(productInfo.Storage),
		OsType:             s.naChecker(productInfo.OperatingSystem),
		ProductDescription: s.naChecker(productInfo.Description),
		PricePolicy:        constant.OnDemand,
		LastUpdatedAt:      time.Now(),
		ImageName:          param.Image,
	}
}

// awsTermsPrices reads the AWS price list structure (terms.OnDemand.*.priceDimensions.*.pricePerUnit.USD).
// found reports whether the structure held any price, even if every price was filtered out by the validator,
// so that a filtered AWS item does not fall through to the next source.
func (s *SpiderPriceCollector) awsTermsPrices(param RecommendSpecParam, base EstimateCostInfo, cspPriceInfo interface{}) (ret []*EstimateCostInfo, found bool) {
	cspInfo, ok := cspPriceInfo.(map[string]interface{})
	if !ok {
		return nil, false
	}
	terms, ok := cspInfo["terms"].(map[string]interface{})
	if !ok {
		return nil, false
	}
	onDemand, ok := terms["OnDemand"].(map[string]interface{})
	if !ok {
		return nil, false
	}

	for _, termData := range onDemand {
		termMap, ok := termData.(map[string]interface{})
		if !ok {
			continue
		}
		priceDimensions, ok := termMap["priceDimensions"].(map[string]interface{})
		if !ok {
			continue
		}

		for _, dimensionData := range priceDimensions {
			dimensionMap, ok := dimensionData.(map[string]interface{})
			if !ok {
				continue
			}
			pricePerUnit, ok := dimensionMap["pricePerUnit"].(map[string]interface{})
			if !ok {
				continue
			}
			usdPrice, ok := pricePerUnit["USD"].(string)
			if !ok {
				continue
			}

			found = true
			description, _ := dimensionMap["description"].(string)

			pi := base
			pi.OriginalPricePolicy = "OnDemand"
			pi.Price = usdPrice
			pi.Currency = constant.USD
			pi.Unit = constant.PerHour
			pi.OriginalUnit = "Hrs"
			pi.OriginalCurrency = "USD"
			pi.PriceDescription = description
			pi.CalculatedMonthlyPrice = s.calculatePrice(usdPrice, constant.PerHour)

			if !priceValidator[param.ProviderName](&pi) {
				log.Debug().Msgf("Price info filtered out by priceValidator: %s %s %s - PriceDescription: %s", pi.ProviderName, pi.RegionName, pi.InstanceType, pi.PriceDescription)
				continue
			}

			ret = append(ret, &pi)
		}
	}

	return ret, found
}

// onDemandPrice reads PriceInfo.OnDemand, the single on-demand price cb-spider v0.11.5+ sends for every provider.
func (s *SpiderPriceCollector) onDemandPrice(param RecommendSpecParam, base EstimateCostInfo, onDemand spider.OnDemandRes) *EstimateCostInfo {
	if onDemand.Price == "" {
		return nil
	}

	convertedPrice, err := strconv.ParseFloat(onDemand.Price, 64)
	if err != nil {
		log.Warn().Msgf("OnDemand price parse error: %s", err)
		return nil
	}
	if convertedPrice == float64(0) {
		log.Warn().Msg("OnDemand price empty (=0)")
		return nil
	}

	pi := base
	pi.OriginalPricePolicy = "OnDemand"
	pi.Price = s.naChecker(onDemand.Price)
	pi.OriginalCurrency = s.naChecker(onDemand.Currency)
	pi.Currency = s.parseCurrency(onDemand.Currency)
	pi.OriginalUnit = s.naChecker(onDemand.Unit)
	pi.Unit = s.parseUnit(pi.OriginalUnit)
	pi.PriceDescription = s.naChecker(onDemand.Description)
	pi.CalculatedMonthlyPrice = s.calculatePrice(pi.Price, pi.Unit)

	if !priceValidator[param.ProviderName](&pi) {
		log.Debug().Msgf("Price info filtered out by priceValidator: %s %s %s - PriceDescription: %s", pi.ProviderName, pi.RegionName, pi.InstanceType, pi.PriceDescription)
		return nil
	}

	return &pi
}

// pricingPolicyPrices reads PriceInfo.PricingPolicies sent by cb-spider < v0.11.5.
func (s *SpiderPriceCollector) pricingPolicyPrices(param RecommendSpecParam, base EstimateCostInfo, policies []spider.PricingPoliciesRes) []*EstimateCostInfo {
	ret := make([]*EstimateCostInfo, 0)

	for k := range policies {
		policy := policies[k]

		convertedPrice, err := strconv.ParseFloat(policy.Price, 64)
		if err != nil {
			log.Warn().Msgf("not allowed for error; %s", err)
			continue
		}
		if convertedPrice == float64(0) {
			log.Warn().Msg("not allowed for empty price")
			continue
		}

		price := s.naChecker(policy.Price)
		if price == "" {
			log.Warn().Msg("not allowed for empty price")
			continue
		}

		priceDescription := s.naChecker(policy.Description)
		if strings.Contains(strings.ToLower(priceDescription), "dedicated") {
			log.Debug().Msgf("not allowed for dedicated instance hour; %s", priceDescription)
			continue
		}

		pi := base
		pi.OriginalPricePolicy = s.naChecker(policy.PricingPolicy)
		pi.Price = price
		pi.OriginalCurrency = s.naChecker(policy.Currency)
		pi.Currency = s.parseCurrency(policy.Currency)
		pi.OriginalUnit = s.naChecker(policy.Unit)
		pi.Unit = s.parseUnit(pi.OriginalUnit)
		pi.PriceDescription = priceDescription
		pi.CalculatedMonthlyPrice = s.calculatePrice(price, pi.Unit)

		if !priceValidator[param.ProviderName](&pi) {
			log.Debug().Msgf("Price info filtered out by priceValidator: %s %s %s - PriceDescription: %s", pi.ProviderName, pi.RegionName, pi.InstanceType, pi.PriceDescription)
			continue
		}

		ret = append(ret, &pi)
	}

	return ret
}

// resolveRegion keeps the requested region when cb-spider leaves VMSpecInfo.Region empty (e.g. GCP).
func resolveRegion(spec *spider.VMSpecInfoRes, requested string) string {
	if spec != nil && strings.TrimSpace(spec.Region) != "" {
		return spec.Region
	}
	return requested
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
		// Remove instanceType filter as it doesn't work properly in CB-Spider v0.11.5
		// CM-Ant receives all instance types and filters them
		// {
		//	Key:   "instanceType",
		//	Value: param.InstanceType,
		// },
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

// mibToGiB converts cb-spider MemSizeMiB (a unitless MiB number) to GiB,
// rounded to one decimal place without a trailing zero (8192 -> "8", 1740.8 -> "1.7").
func (s *SpiderPriceCollector) mibToGiB(input string) (string, constant.MemoryUnit) {
	if input == "" {
		return "", ""
	}

	mib, err := strconv.ParseFloat(strings.TrimSpace(input), 64)
	if err != nil {
		log.Warn().Msgf("error while converting memory from MiB : %s", err.Error())
		return "", ""
	}

	gib := strconv.FormatFloat(math.Round(mib/1024*10)/10, 'f', -1, 64)
	return gib, constant.GIB
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
