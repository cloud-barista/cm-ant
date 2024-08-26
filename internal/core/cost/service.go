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

func (c *CostService) GetPriceInfo(param GetPriceInfoParam) (AllPriceInfoResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	param.TimeStandard = time.Now().AddDate(0, 0, -7).Truncate(24 * time.Hour)
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
				RegionName:             v.RegionName,
				InstanceType:           v.InstanceType,
				ZoneName:               v.ZoneName,
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
			pil = append(pil, r)
		}

		res.InfoSource = "db"
		res.PriceInfoList = pil
		res.ResultCount = int64(len(pil))

		return res, nil
	}

	resList, err := c.priceCollector.GetPriceInfos(ctx, param)

	if err != nil {
		if strings.Contains(err.Error(), "you don't have any permission") {
			return res, fmt.Errorf("you don't have permission to query the price for %s", param.ProviderName)
		}
		return res, err
	}

	if len(resList) > 0 {
		err := c.costRepo.InsertAllResult(ctx, param, resList)
		if err != nil {
			return res, nil
		}

		for _, v := range resList {
			r := PriceInfoResult{
				ID:                     v.ID,
				ProviderName:           v.ProviderName,
				RegionName:             v.RegionName,
				InstanceType:           v.InstanceType,
				ZoneName:               v.ZoneName,
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
