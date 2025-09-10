package cost

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/rs/zerolog/log"
)

type CostService struct {
	costRepo       *CostRepository
	priceCollector PriceCollector
	costCollector  CostCollector
}

func NewCostService(costRepo *CostRepository, priceCollector PriceCollector, costCollector CostCollector) *CostService {
	return &CostService{
		costRepo:       costRepo,
		priceCollector: priceCollector,
		costCollector:  costCollector,
	}
}

func (c *CostService) Readyz() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sqlDB, err := c.costRepo.db.DB()
	if err != nil {
		return err
	}

	err = sqlDB.Ping()
	if err != nil {
		return err
	}

	err = c.costCollector.Readyz(ctx)
	if err != nil {
		return err
	}

	err = c.priceCollector.Readyz(ctx)
	if err != nil {
		return err
	}

	return nil
}

var estimateCostUpdateLockMap sync.Map

func (c *CostService) UpdateAndGetEstimateCost(param UpdateAndGetEstimateCostParam) (EstimateCostResults, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Set TimeStandard to far past to include all data
	param.TimeStandard = time.Now().AddDate(0, 0, -365).Truncate(24 * time.Hour) // Set to 1 year ago

	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []EsimateCostSpecResults
	var errList []error
	var esimateCostSpecResult EstimateCostResults

	log.Info().Msgf("Fetching estimate cost info for spec: %+v", param)

	fail := func(msgFormat string, err error, p RecommendSpecParam) {
		mu.Lock()
		errList = append(errList, err)
		mu.Unlock()
		log.Error().Msgf(msgFormat, p, err)
	}

	for _, v := range param.RecommendSpecs {
		wg.Add(1)
		go func(p RecommendSpecParam) {
			defer wg.Done()

			// memory lock
			rl, _ := estimateCostUpdateLockMap.LoadOrStore(p.Hash(), &sync.Mutex{})
			lock := rl.(*sync.Mutex)

			lock.Lock()
			defer lock.Unlock()

			var err error
			var estimateCostInfos EstimateCostInfos
			possibleFetch := true

			if p.ProviderName == "ibm" || p.ProviderName == "azure" {
				r, err := c.costRepo.GetMatchingEstimateCostWithoutTypeTx(ctx, v, param.TimeStandard, param.PricePolicy)
				if err != nil {
					fail("Error fetching estimate cost info: %v; %s", err, p)
					return
				}

				// ibm and azure price fetching filter is not working so if user put never exist instance type,
				// it always fetch price data from collector.
				// check and matching with database values
				if len(r) != 0 {
					temp := make([]*EstimateCostInfo, 0)

					for i := range r {
						val := r[i]

						if val == nil {
							log.Warn().Msg("estimate cost info value is nil")
							continue
						}

						if strings.EqualFold(v.InstanceType, p.InstanceType) {
							temp = append(temp, val)
						}
					}

					if len(temp) > 0 {
						estimateCostInfos = temp
					} else {
						possibleFetch = false
					}
				}
			} else if strings.Contains(p.ProviderName, "ncp") {
				log.Warn().Msgf("%s provider's cost is bypassing", p.ProviderName)
				return
			} else {
				estimateCostInfos, err = c.costRepo.GetMatchingEstimateCostTx(ctx, p, param.TimeStandard, param.PricePolicy)
			}

			if err != nil {
				fail("Error fetching estimate cost info for spec: %v; %s", err, p)
				return
			}

			if len(estimateCostInfos) == 0 || possibleFetch {
				log.Info().Msgf("No matching estimate cost found from database for spec: %+v, fetching from price collector", p)

				resList, err := c.priceCollector.FetchPriceInfos(ctx, p)
				if err != nil {
					fail("Error retrieving estimate cost info spec: %v; %s", fmt.Errorf("error retrieving estimate cost info for %+v: %w", p, err), p)
					return
				}

				if len(resList) > 0 {
					log.Info().Msgf("Inserting fetched estimate cost info results for spec: %+v", p)

					err = c.costRepo.BatchInsertAllEstimateCostResultTx(ctx, resList)
					if err != nil {
						fail("Error batch inserting estimate cost info spec: %v; %s", fmt.Errorf("error batch inserting results for %+v: %w", p, err), p)
						return
					}
				}
				estimateCostInfos = resList
			}

			res := EsimateCostSpecResults{
				ProviderName:                  p.ProviderName,
				RegionName:                    p.RegionName,
				InstanceType:                  p.InstanceType,
				ImageName:                     p.Image,
				EstimateCostSpecDetailResults: make([]EstimateCostSpecDetailResult, 0),
			}

			if len(estimateCostInfos) > 0 {
				minPrice := estimateCostInfos[0].CalculatedMonthlyPrice
				maxPrice := estimateCostInfos[0].CalculatedMonthlyPrice

				for _, va := range estimateCostInfos {

					if !strings.EqualFold(v.InstanceType, p.InstanceType) {
						log.Warn().Msgf("%s instance type is not matching with provided condition %s", va.InstanceType, p.InstanceType)
						continue
					}
					calculatedPrice := va.CalculatedMonthlyPrice
					log.Info().Msgf("Price calculated for spec %+v: %f", p, calculatedPrice)

					if calculatedPrice < minPrice {
						minPrice = calculatedPrice
					}
					if calculatedPrice > maxPrice {
						maxPrice = calculatedPrice
					}

					specDetail := EstimateCostSpecDetailResult{
						ID:                     va.ID,
						VCpu:                   va.VCpu,
						Memory:                 fmt.Sprintf("%s %s", va.Memory, va.MemoryUnit),
						Storage:                va.Storage,
						OsType:                 va.OsType,
						ProductDescription:     va.ProductDescription,
						OriginalPricePolicy:    va.OriginalPricePolicy,
						PricePolicy:            va.PricePolicy,
						Unit:                   va.Unit,
						Currency:               va.Currency,
						Price:                  va.Price,
						CalculatedMonthlyPrice: calculatedPrice,
						PriceDescription:       va.PriceDescription,
						LastUpdatedAt:          va.LastUpdatedAt,
					}

					res.SpecMinMonthlyPrice = minPrice
					res.SpecMaxMonthlyPrice = maxPrice
					res.EstimateCostSpecDetailResults = append(res.EstimateCostSpecDetailResults, specDetail)
				}

				if len(res.EstimateCostSpecDetailResults) > 0 {
					sort.Slice(res.EstimateCostSpecDetailResults, func(i, j int) bool {
						return res.EstimateCostSpecDetailResults[i].CalculatedMonthlyPrice < res.EstimateCostSpecDetailResults[j].CalculatedMonthlyPrice
					})
				}

				mu.Lock()
				results = append(results, res)
				mu.Unlock()
				log.Info().Msgf("Successfully calculated cost for spec: %+v", param)
			}

		}(v)
	}
	wg.Wait()

	if len(errList) > 0 {
		return esimateCostSpecResult, fmt.Errorf("errors occurred during processing: %v", errList)
	}

	if len(results) > 0 {
		esimateCostSpecResult.EsimateCostSpecResults = results
	}

	return esimateCostSpecResult, nil
}

func (c *CostService) GetEstimateCost(param GetEstimateCostParam) (EstimateCostInfoResults, error) {
	log.Info().Msgf("=== GetEstimateCost START ===")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Set TimeStandard to far past to include all data
	param.TimeStandard = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC) // Set to year 2000
	param.PricePolicy = constant.OnDemand

	log.Info().Msgf("GetEstimateCost called with TimeStandard: %s", param.TimeStandard)

	var res EstimateCostInfoResults

	estimateCostInfos, totalCount, err := c.costRepo.GetMatchingEstimateCostInfosTx(ctx, param)
	if err != nil {
		return res, err
	}

	priceInfoList := make([]EstimateCostInfoResult, 0)

	if len(estimateCostInfos) > 0 {
		for _, v := range estimateCostInfos {
			result := EstimateCostInfoResult{
				ID:                     v.ID,
				ProviderName:           v.ProviderName,
				RegionName:             v.RegionName,
				InstanceType:           v.InstanceType,
				VCpu:                   v.VCpu,
				Memory:                 fmt.Sprintf("%s %s", v.Memory, v.MemoryUnit),
				Storage:                v.Storage,
				OsType:                 v.OsType,
				ProductDescription:     v.ProductDescription,
				OriginalPricePolicy:    v.OriginalPricePolicy,
				PricePolicy:            v.PricePolicy,
				Unit:                   v.Unit,
				Currency:               v.Currency,
				Price:                  v.Price,
				CalculatedMonthlyPrice: v.CalculatedMonthlyPrice,
				PriceDescription:       v.PriceDescription,
				LastUpdatedAt:          v.UpdatedAt,
			}
			priceInfoList = append(priceInfoList, result)
		}

		res.EstimateCostInfoResult = priceInfoList
		res.ResultCount = int64(totalCount)

		return res, nil
	}

	return res, nil
}

func (c *CostService) UpdateEstimateForecastCost(param UpdateEstimateForecastCostParam) (UpdateEstimateForecastCostInfoResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	var updateEstimateForecastCostInfoResult UpdateEstimateForecastCostInfoResult

	r, err := c.costCollector.UpdateEstimateForecastCost(ctx, param)
	if err != nil {
		return updateEstimateForecastCostInfoResult, err
	}

	updateEstimateForecastCostInfoResult.FetchedDataCount = int64(len(r))

	var updatedCount int64
	var insertedCount int64

	for _, costInfo := range r {
		u, i, err := c.costRepo.UpsertCostInfo(ctx, costInfo)
		if err != nil {
			log.Error().Msgf("upsert error: %+v", costInfo)
		}

		updatedCount += u
		insertedCount += i
	}

	log.Info().Msgf("updated count: %d; inserted count : %d", updatedCount, insertedCount)

	updateEstimateForecastCostInfoResult.UpdatedDataCount = updatedCount
	updateEstimateForecastCostInfoResult.InsertedDataCount = insertedCount

	return updateEstimateForecastCostInfoResult, nil
}

func (c *CostService) GetEstimateForecastCostInfos(param GetEstimateForecastCostParam) (GetEstimateForecastCostInfoResults, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	res := GetEstimateForecastCostInfoResults{}

	r, totalCount, err := c.costRepo.GetEstimateForecastCostInfosTx(ctx, param)
	if err != nil {
		return res, err
	}

	res.GetEstimateForecastCostInfoResults = r
	res.ResultCount = totalCount

	return res, nil
}

var (
	ErrRequestResourceEmpty    = errors.New("cost request info is not enough")
	ErrCostResultEmpty         = errors.New("cost information does not exist")
	ErrCostResultFormatInvalid = errors.New("cost result does not matching with interface")
)

func (c *CostService) UpdateEstimateForecastCostRaw(param UpdateEstimateForecastCostRawParam) (UpdateEstimateForecastCostInfoResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	var updateCostInfoResult UpdateEstimateForecastCostInfoResult

	r, err := c.costCollector.GetCostInfos(ctx, param)
	if err != nil {
		return updateCostInfoResult, err
	}

	updateCostInfoResult.FetchedDataCount = int64(len(r))

	var updatedCount int64
	var insertedCount int64

	for _, costInfo := range r {
		u, i, err := c.costRepo.UpsertCostInfo(ctx, costInfo)
		if err != nil {
			log.Error().Msgf("upsert error: %+v", costInfo)
		}

		updatedCount += u
		insertedCount += i
	}
	log.Info().Msgf("updated count: %d; inserted count : %d", updatedCount, insertedCount)

	updateCostInfoResult.UpdatedDataCount = updatedCount
	updateCostInfoResult.InsertedDataCount = insertedCount

	return updateCostInfoResult, nil
}
