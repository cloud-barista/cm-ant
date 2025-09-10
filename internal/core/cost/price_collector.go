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

	// Log CB-Spider response results
	log.Info().Msgf("CB-Spider result for %s %s %s: CloudPriceList=%v, DirectPriceList=%v", param.ProviderName, param.RegionName, param.InstanceType, result.CloudPriceList != nil, result.PriceList != nil)
	if result.CloudPriceList != nil {
		log.Info().Msgf("CloudPriceList length: %d", len(result.CloudPriceList))
	}
	if result.PriceList != nil {
		log.Info().Msgf("Direct PriceList length: %d", len(result.PriceList))
	}

	// Debug: Check actual response structure
	log.Info().Msgf("Response structure debug - CloudName: %s, RegionName: %s", result.CloudName, result.RegionName)

	createdPriceInfo := make([]*EstimateCostInfo, 0)

	// Process v0.11.5 API response structure (when PriceList is directly available)
	if result.PriceList != nil && len(result.PriceList) > 0 {
		log.Info().Msgf("Processing v0.11.5 API response structure with %d price items", len(result.PriceList))

		// Log detailed information for first few items
		for i := 0; i < len(result.PriceList) && i < 3; i++ {
			pl := result.PriceList[i]
			log.Info().Msgf("Sample PriceList[%d]: InstanceType=%s, vCpu=%s, Memory=%s",
				i, pl.ProductInfo.InstanceType, pl.ProductInfo.Vcpu, pl.ProductInfo.Memory)
		}

		for j := range result.PriceList {
			pl := result.PriceList[j]
			log.Info().Msgf("Processing Direct PriceList[%d]: InstanceType=%s", j, pl.ProductInfo.InstanceType)

			productInfo := pl.ProductInfo

			// Extract vCpu and Memory from v0.11.5 structure
			var vCpu, originalMemory, instanceType string

			if productInfo.VMSpecInfo != nil {
				// v0.11.5 new structure
				vCpu = s.naChecker(productInfo.VMSpecInfo.VCpu.Count)
				originalMemory = s.naChecker(productInfo.VMSpecInfo.MemSizeMiB)
				instanceType = s.naChecker(productInfo.VMSpecInfo.Name)
				log.Info().Msgf("Using v0.11.5 VMSpecInfo: vCpu=%s, Memory=%s, InstanceType=%s", vCpu, originalMemory, instanceType)
			} else {
				// v0.10.0 legacy structure
				vCpu = s.naChecker(productInfo.Vcpu)
				originalMemory = s.naChecker(productInfo.Memory)
				instanceType = s.naChecker(productInfo.InstanceType)
				log.Info().Msgf("Using v0.10.0 ProductInfo: vCpu=%s, Memory=%s, InstanceType=%s", vCpu, originalMemory, instanceType)
			}

			log.Info().Msgf("ProductInfo: vCpu=%s, Memory=%s, InstanceType=%s, RequestedType=%s", vCpu, originalMemory, instanceType, param.InstanceType)

			// Debug logging
			if strings.EqualFold(instanceType, param.InstanceType) {
				log.Error().Msgf("Match found: got=%s, requested=%s, vCpu=%s, originalMemory=%s", instanceType, param.InstanceType, vCpu, originalMemory)
			}

			// Check instance type matching
			if !strings.EqualFold(instanceType, param.InstanceType) {
				log.Info().Msgf("Skipping due to instance type mismatch: got=%s, requested=%s", instanceType, param.InstanceType)
				continue
			}

			if vCpu == "" || originalMemory == "" {
				log.Info().Msgf("Skipping due to empty vCpu or Memory: vCpu=%s, Memory=%s", vCpu, originalMemory)
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

			// Add priceInfo debug logs
			log.Info().Msgf("PriceInfo debug for %s: PricingPolicies=%v, CSPPriceInfo=%v", instanceType, priceInfo.PricingPolicies != nil, priceInfo.CSPPriceInfo != nil)
			if priceInfo.PricingPolicies != nil {
				log.Info().Msgf("PricingPolicies count: %d", len(priceInfo.PricingPolicies))
				for idx, policy := range priceInfo.PricingPolicies {
					log.Info().Msgf("PricingPolicy[%d]: Policy=%s, Price=%s, Currency=%s, Unit=%s, Description=%s",
						idx, policy.PricingPolicy, policy.Price, policy.Currency, policy.Unit, policy.Description)
				}
			} else {
				log.Warn().Msgf("PricingPolicies is nil for instance type: %s", instanceType)
			}

			// Add CSPPriceInfo debug logs
			if priceInfo.CSPPriceInfo != nil {
				log.Info().Msgf("CSPPriceInfo exists for %s: %+v", instanceType, priceInfo.CSPPriceInfo)
			} else {
				log.Warn().Msgf("CSPPriceInfo is nil for instance type: %s", instanceType)
			}

			// Extract price information from v0.11.5 CSPPriceInfo
			if priceInfo.CSPPriceInfo != nil {
				log.Info().Msgf("Processing CSPPriceInfo for %s", instanceType)

				// CSPPriceInfo is in map[string]interface{} format
				if cspInfo, ok := priceInfo.CSPPriceInfo.(map[string]interface{}); ok {
					if terms, ok := cspInfo["terms"].(map[string]interface{}); ok {
						if onDemand, ok := terms["OnDemand"].(map[string]interface{}); ok {
							// Process OnDemand price information
							for sku, termData := range onDemand {
								if termMap, ok := termData.(map[string]interface{}); ok {
									if priceDimensions, ok := termMap["priceDimensions"].(map[string]interface{}); ok {
										for rateCode, dimensionData := range priceDimensions {
											if dimensionMap, ok := dimensionData.(map[string]interface{}); ok {
												if pricePerUnit, ok := dimensionMap["pricePerUnit"].(map[string]interface{}); ok {
													if usdPrice, ok := pricePerUnit["USD"].(string); ok {
														description := ""
														if desc, ok := dimensionMap["description"].(string); ok {
															description = desc
														}

														log.Info().Msgf("Found OnDemand price: SKU=%s, RateCode=%s, Price=%s, Description=%s", sku, rateCode, usdPrice, description)

														// Create price information
														regionName := param.RegionName // default value
														if productInfo.VMSpecInfo != nil {
															regionName = productInfo.VMSpecInfo.Region // Use VMSpecInfo.Region in CB-Spider v0.11.5 to store accurate region information
														}

														pi := EstimateCostInfo{
															ProviderName:           param.ProviderName,
															RegionName:             regionName,
															InstanceType:           instanceType,
															ZoneName:               zoneName,
															VCpu:                   vCpu,
															OriginalMemory:         originalMemory,
															Memory:                 memory,
															MemoryUnit:             memoryUnit,
															Storage:                storage,
															OsType:                 osType,
															ProductDescription:     productDescription,
															OriginalPricePolicy:    "OnDemand",
															PricePolicy:            constant.OnDemand,
															Price:                  usdPrice,
															Currency:               constant.USD,
															Unit:                   constant.PerHour,
															OriginalUnit:           "Hrs",
															OriginalCurrency:       "USD",
															PriceDescription:       description,
															CalculatedMonthlyPrice: s.calculatePrice(usdPrice, constant.PerHour),
															LastUpdatedAt:          time.Now(),
															ImageName:              param.Image,
														}

														log.Info().Msgf("Before priceValidator check: Provider=%s, PriceDescription=%s, Price=%s", pi.ProviderName, pi.PriceDescription, pi.Price)
														if !priceValidator[param.ProviderName](&pi) {
															log.Info().Msgf("Price info filtered out by priceValidator: %s %s %s - PriceDescription: %s", pi.ProviderName, pi.RegionName, pi.InstanceType, pi.PriceDescription)
															continue
														}

														createdPriceInfo = append(createdPriceInfo, &pi)
														log.Info().Msgf("Added price info (CSPPriceInfo): %s %s %s - $%s", pi.ProviderName, pi.RegionName, pi.InstanceType, pi.Price)
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			} else if priceInfo.PricingPolicies != nil {
				// Process existing v0.10.0 PricingPolicies
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

					regionName := param.RegionName // default value
					if productInfo.VMSpecInfo != nil {
						regionName = productInfo.VMSpecInfo.Region // Use VMSpecInfo.Region in CB-Spider v0.11.5 to store accurate region information
					}

					pi := EstimateCostInfo{
						ProviderName:           param.ProviderName,
						RegionName:             regionName,
						InstanceType:           instanceType,
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

					log.Info().Msgf("Before priceValidator check: Provider=%s, PriceDescription=%s, Price=%s", pi.ProviderName, pi.PriceDescription, pi.Price)
					if !priceValidator[param.ProviderName](&pi) {
						log.Info().Msgf("Price info filtered out by priceValidator: %s %s %s - PriceDescription: %s", pi.ProviderName, pi.RegionName, pi.InstanceType, pi.PriceDescription)
						continue
					}

					createdPriceInfo = append(createdPriceInfo, &pi)
					log.Info().Msgf("Added price info (v0.10.0): %s %s %s - $%s", pi.ProviderName, pi.RegionName, pi.InstanceType, pi.Price)
				}
			}
		}
	}

	// Process existing v0.10.0 API response structure (when CloudPriceList is available)
	if result.CloudPriceList != nil {
		log.Info().Msgf("Processing CloudPriceList with %d items", len(result.CloudPriceList))
		for i := range result.CloudPriceList {
			p := result.CloudPriceList[i]
			log.Info().Msgf("Processing CloudPrice[%d]: CloudName=%s, PriceList count=%d", i, p.CloudName, len(p.PriceList))

			if p.PriceList != nil {
				for j := range p.PriceList {

					pl := p.PriceList[j]
					log.Info().Msgf("Processing PriceList[%d]: InstanceType=%s", j, pl.ProductInfo.InstanceType)

					productInfo := pl.ProductInfo
					vCpu := s.naChecker(productInfo.Vcpu)
					originalMemory := s.naChecker(productInfo.Memory)

					log.Info().Msgf("ProductInfo: vCpu=%s, Memory=%s, InstanceType=%s", vCpu, originalMemory, productInfo.InstanceType)

					if vCpu == "" || originalMemory == "" {
						log.Info().Msgf("Skipping due to empty vCpu or Memory: vCpu=%s, Memory=%s", vCpu, originalMemory)
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

							regionName := param.RegionName // default value
							if productInfo.VMSpecInfo != nil {
								regionName = productInfo.VMSpecInfo.Region // Use VMSpecInfo.Region in CB-Spider v0.11.5 to store accurate region information
							}

							pi := EstimateCostInfo{
								ProviderName:           param.ProviderName,
								RegionName:             regionName,
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
							log.Info().Msgf("Added price info: %s %s %s - $%s", pi.ProviderName, pi.RegionName, pi.InstanceType, pi.Price)
						}
					}

				}
			}
		}
	}

	log.Info().Msgf("Total created price info count: %d", len(createdPriceInfo))

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
