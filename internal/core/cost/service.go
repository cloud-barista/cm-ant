package cost

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/cloud-barista/cm-ant/internal/utils"
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

var forecastUpdateLockMap sync.Map

func (c *CostService) EstimateForecastCost(param EstimateForecastCostParam) (EstimateForecastCostResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []EsimateForecastCostSpecResult
	var errList []error
	var esimateForecastCostSpecResult EstimateForecastCostResult

	utils.LogInfof("Fetching price information for spec: %+v", param)

	for _, v := range param.RecommendSpecs {
		wg.Add(1)
		go func(p RecommendSpecParam) {
			defer wg.Done()

			// memory lock
			rl, _ := forecastUpdateLockMap.LoadOrStore(p.Hash(), &sync.Mutex{})
			lock := rl.(*sync.Mutex)

			lock.Lock()
			defer lock.Unlock()

			priceInfos, err := c.costRepo.GetMatchingForecastCost(ctx, v, param.TimeStandard, param.PricePolicy)
			if err != nil {
				mu.Lock()
				errList = append(errList, err)
				mu.Unlock()
				utils.LogErrorf("Error fetching price info for spec %+v: %v", v, err)

				return

			}

			if len(priceInfos) == 0 {
				utils.LogInfof("No matching forecast cost found for spec: %+v, fetching from price collector", v)

				resList, err := c.priceCollector.FetchPriceInfos(ctx, v)
				if err != nil {
					mu.Lock()
					errList = append(errList, fmt.Errorf("error retrieving prices for %+v: %w", v, err))
					mu.Unlock()
					return
				}

				if len(resList) > 0 {
					utils.LogInfof("Inserting fetched price results for spec: %+v", v)

					err = c.costRepo.BatchInsertAllForecastCostResult(ctx, resList)
					if err != nil {
						mu.Lock()
						errList = append(errList, fmt.Errorf("error batch inserting results for %+v: %w", v, err))
						mu.Unlock()
						return
					}
				}
				priceInfos = resList
			}

			if len(priceInfos) > 0 {

				minPrice := float64(math.MaxFloat64)
				maxPrice := float64(math.SmallestNonzeroFloat64)

				res := EsimateForecastCostSpecResult{
					ProviderName:                          v.ProviderName,
					RegionName:                            v.RegionName,
					InstanceType:                          v.InstanceType,
					ImageName:                             v.Image,
					EstimateForecastCostSpecDetailResults: make([]EstimateForecastCostSpecDetailResult, 0),
				}

				for _, v := range priceInfos {

					calculatedPrice := v.CalculatedMonthlyPrice
					utils.LogInfof("Price calculated for spec %+v: %f", v, calculatedPrice)

					if calculatedPrice < minPrice {
						minPrice = calculatedPrice
					}
					if calculatedPrice > maxPrice {
						maxPrice = calculatedPrice
					}

					specDetail := EstimateForecastCostSpecDetailResult{
						ID:                     v.ID,
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
						CalculatedMonthlyPrice: calculatedPrice,
						PriceDescription:       v.PriceDescription,
						LastUpdatedAt:          v.LastUpdatedAt,
					}

					res.SpecMinMonthlyPrice = minPrice
					res.SpecMaxMonthlyPrice = maxPrice
					res.EstimateForecastCostSpecDetailResults = append(res.EstimateForecastCostSpecDetailResults, specDetail)
				}

				mu.Lock()
				results = append(results, res)
				mu.Unlock()
				utils.LogInfof("Successfully calculated forecast cost for spec: %+v", param)
			}

		}(v)
	}
	wg.Wait()

	if len(errList) > 0 {
		return esimateForecastCostSpecResult, fmt.Errorf("errors occurred during processing: %v", errList)
	}

	if len(results) > 0 {
		esimateForecastCostSpecResult.EsimateForecastCostSpecResults = results

		for _, v := range results {
			esimateForecastCostSpecResult.TotalMinMonthlyPrice += v.SpecMinMonthlyPrice
			esimateForecastCostSpecResult.TotalMaxMonthlyPrice += v.SpecMaxMonthlyPrice
		}
		utils.LogInfof("Total min monthly price: %f, Total max monthly price: %f", esimateForecastCostSpecResult.TotalMinMonthlyPrice, esimateForecastCostSpecResult.TotalMaxMonthlyPrice)

	}

	return esimateForecastCostSpecResult, nil
}

func (c *CostService) UpdatePriceInfos(param UpdatePriceInfosParam) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	param.TimeStandard = time.Now().AddDate(0, 0, -7).Truncate(24 * time.Hour)
	param.PricePolicy = constant.OnDemand

	count, err := c.costRepo.CountMatchingPriceInfoList(ctx, param)
	if err != nil {
		return err
	}

	if count <= int64(0) {
		resList, err := c.priceCollector.GetPriceInfos(ctx, param)

		if err != nil {
			if strings.Contains(err.Error(), "you don't have any permission") {
				return fmt.Errorf("you don't have permission to query the price for %s", param.ProviderName)
			}
			return err
		}

		if len(resList) > 0 {
			err := c.costRepo.BatchInsertAllForecastCostResult(ctx, resList)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *CostService) GetPriceInfos(param GetPriceInfosParam) (AllPriceInfoResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	param.TimeStandard = time.Now().AddDate(0, 0, -7).Truncate(24 * time.Hour)
	param.PricePolicy = constant.OnDemand

	var res AllPriceInfoResult

	priceInfos, err := c.costRepo.GetAllMatchingPriceInfoList(ctx, param)
	if err != nil {
		return res, err
	}

	priceInfoList := make([]PriceInfoResult, 0)

	if len(priceInfos) > 0 {
		for _, v := range priceInfos {
			result := PriceInfoResult{
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

		res.PriceInfoList = priceInfoList
		res.ResultCount = int64(len(priceInfoList))

		return res, nil
	}

	return res, nil
}

var (
	ErrRequestResourceEmpty    = errors.New("cost request info is not enough")
	ErrCostResultEmpty         = errors.New("cost information does not exist")
	ErrCostResultFormatInvalid = errors.New("cost result does not matching with interface")
)

func (c *CostService) UpdateCostInfo(param UpdateCostInfoParam) (UpdateCostInfoResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	var updateCostInfoResult UpdateCostInfoResult

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
			utils.LogErrorf("upsert error: %+v", costInfo)
		}

		updatedCount += u
		insertedCount += i
	}
	utils.LogInfof("updated count: %d; inserted count : %d", updatedCount, insertedCount)

	updateCostInfoResult.UpdatedDataCount = updatedCount
	updateCostInfoResult.InsertedDataCount = insertedCount

	return updateCostInfoResult, nil
}

func (c *CostService) GetCostInfos(param GetCostInfoParam) ([]GetCostInfoResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	r, err := c.costRepo.GetCostInfoWithFilter(ctx, param)
	if err != nil {
		return nil, err
	}
	return r, nil
}
