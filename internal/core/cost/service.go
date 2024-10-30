package cost

import (
	"context"
	"errors"
	"fmt"
	"math"
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

var estimateCostUpdateLockMap sync.Map

func (c *CostService) UpdateAndGetEstimateCost(param UpdateAndGetEstimateCostParam) (EstimateCostResults, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []EsimateCostSpecResults
	var errList []error
	var esimateCostSpecResult EstimateCostResults

	utils.LogInfof("Fetching estimate cost info for spec: %+v", param)

	for _, v := range param.RecommendSpecs {
		wg.Add(1)
		go func(p RecommendSpecParam) {
			defer wg.Done()

			// memory lock
			rl, _ := estimateCostUpdateLockMap.LoadOrStore(p.Hash(), &sync.Mutex{})
			lock := rl.(*sync.Mutex)

			lock.Lock()
			defer lock.Unlock()

			estimateCostInfos, err := c.costRepo.GetMatchingEstimateCostTx(ctx, v, param.TimeStandard, param.PricePolicy)
			if err != nil {
				mu.Lock()
				errList = append(errList, err)
				mu.Unlock()
				utils.LogErrorf("Error fetching estimate cost info for spec %+v: %v", v, err)

				return
			}

			if len(estimateCostInfos) == 0 {
				utils.LogInfof("No matching estimate cost found for spec: %+v, fetching from price collector", v)

				resList, err := c.priceCollector.FetchPriceInfos(ctx, v)
				if err != nil {
					mu.Lock()
					errList = append(errList, fmt.Errorf("error retrieving estimate cost info for %+v: %w", v, err))
					mu.Unlock()
					return
				}

				if len(resList) > 0 {
					utils.LogInfof("Inserting fetched estimate cost info results for spec: %+v", v)

					err = c.costRepo.BatchInsertAllEstimateCostResultTx(ctx, resList)
					if err != nil {
						mu.Lock()
						errList = append(errList, fmt.Errorf("error batch inserting results for %+v: %w", v, err))
						mu.Unlock()
						return
					}
				}
				estimateCostInfos = resList
			}

			if len(estimateCostInfos) > 0 {

				minPrice := float64(math.MaxFloat64)
				maxPrice := float64(math.SmallestNonzeroFloat64)

				res := EsimateCostSpecResults{
					ProviderName:                  v.ProviderName,
					RegionName:                    v.RegionName,
					InstanceType:                  v.InstanceType,
					ImageName:                     v.Image,
					EstimateCostSpecDetailResults: make([]EstimateCostSpecDetailResult, 0),
				}

				for _, v := range estimateCostInfos {
					calculatedPrice := v.CalculatedMonthlyPrice
					utils.LogInfof("Price calculated for spec %+v: %f", v, calculatedPrice)

					if calculatedPrice < minPrice {
						minPrice = calculatedPrice
					}
					if calculatedPrice > maxPrice {
						maxPrice = calculatedPrice
					}

					specDetail := EstimateCostSpecDetailResult{
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
					res.EstimateCostSpecDetailResults = append(res.EstimateCostSpecDetailResults, specDetail)
				}

				mu.Lock()
				results = append(results, res)
				mu.Unlock()
				utils.LogInfof("Successfully calculated cost for spec: %+v", param)
			}

		}(v)
	}
	wg.Wait()

	if len(errList) > 0 {
		return esimateCostSpecResult, fmt.Errorf("errors occurred during processing: %v", errList)
	}

	if len(results) > 0 {
		esimateCostSpecResult.EsimateCostSpecResults = results

		for _, v := range results {
			esimateCostSpecResult.TotalMinMonthlyPrice += v.SpecMinMonthlyPrice
			esimateCostSpecResult.TotalMaxMonthlyPrice += v.SpecMaxMonthlyPrice
		}
		utils.LogInfof("Total min monthly price: %f, Total max monthly price: %f", esimateCostSpecResult.TotalMinMonthlyPrice, esimateCostSpecResult.TotalMaxMonthlyPrice)

	}

	return esimateCostSpecResult, nil
}

func (c *CostService) GetEstimateCost(param GetEstimateCostParam) (EstimateCostInfoResults, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	param.TimeStandard = time.Now().AddDate(0, 0, -7).Truncate(24 * time.Hour)
	param.PricePolicy = constant.OnDemand

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
			utils.LogErrorf("upsert error: %+v", costInfo)
		}

		updatedCount += u
		insertedCount += i
	}

	utils.LogInfof("updated count: %d; inserted count : %d", updatedCount, insertedCount)

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

func (c *CostService) UpdateCostInfo(param UpdateCostInfoParam) (UpdateEstimateForecastCostInfoResult, error) {
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
