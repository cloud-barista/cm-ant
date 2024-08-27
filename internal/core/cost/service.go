package cost

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
			err := c.costRepo.BatchInsertAllResult(ctx, param, resList)
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
