package cost

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"gorm.io/gorm"
)

type CostRepository struct {
	db *gorm.DB
}

func NewCostRepository(db *gorm.DB) *CostRepository {
	return &CostRepository{
		db: db,
	}
}

func (r *CostRepository) execInTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin transaction error: %w", tx.Error)
	}

	err := fn(tx)
	if err != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return fmt.Errorf("rollback error: %v, original error: %w", rbErr, err)
		}
		return err
	}

	return tx.Commit().Error
}

func (r *CostRepository) GetAllMatchingPriceInfoList(ctx context.Context, param GetPriceInfosParam) (PriceInfos, error) {
	var priceInfoList []*PriceInfo

	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		q := d.Model(&PriceInfo{}).
			Where(
				"price_policy = ? AND updated_at >= ?",
				param.PricePolicy, param.TimeStandard,
			).
			Order("calculated_monthly_price asc").
			Limit(8)

		if param.ProviderName != "" {
			q = q.Where("LOWER(provider_name) = ?", strings.ToLower(param.ProviderName))
		}

		if param.RegionName != "" {
			q = q.Where("LOWER(region_name) = ?", strings.ToLower(param.RegionName))
		}

		if param.InstanceType != "" {
			q = q.Where("LOWER(instance_type) = ?", strings.ToLower(param.InstanceType))
		}

		if param.VCpu != "" {
			q = q.Where("LOWER(v_cpu) = ?", strings.ToLower(param.VCpu))
		}

		if param.Memory != "" {
			q = q.Where("LOWER(memory) = ?", strings.ToLower(param.Memory))
		}

		if param.OsType != "" {
			q = q.Where("LOWER(os_type) = ?", strings.ToLower(param.OsType))
		}

		if err := q.Find(&priceInfoList).Error; err != nil {
			return err
		}

		return nil
	})

	return priceInfoList, err
}

func (r *CostRepository) GetMatchingForecastCost(ctx context.Context, param RecommendSpecParam, timeStandard time.Time, pricePolicy constant.PricePolicy) (PriceInfos, error) {
	var priceInfos []*PriceInfo

	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		q := d.Model(&PriceInfo{}).
			Where(
				"LOWER(provider_name) = ? AND LOWER(region_name) = ? AND instance_type  = ? AND image_name  = ? AND price_policy = ? AND last_updated_at >= ?",
				strings.ToLower(param.ProviderName),
				strings.ToLower(param.RegionName),
				strings.ToLower(param.InstanceType),
				strings.ToLower(param.Image),
				pricePolicy,
				timeStandard,
			)

		if err := q.Find(&priceInfos).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return priceInfos, nil
}

func (r *CostRepository) CountMatchingPriceInfoList(ctx context.Context, param UpdatePriceInfosParam) (int64, error) {
	var totalCount int64

	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		q := d.Model(&PriceInfo{}).
			Where(
				"LOWER(provider_name) = ? AND LOWER(region_name) = ? AND price_policy = ? AND updated_at >= ?",
				strings.ToLower(param.ProviderName), strings.ToLower(param.RegionName), param.PricePolicy, param.TimeStandard,
			)

		if param.InstanceType != "" {
			q = q.Where("LOWER(instance_type) = ?", strings.ToLower(param.InstanceType))
		}

		return q.Count(&totalCount).Error
	})

	return totalCount, err

}

func (r *CostRepository) BatchInsertAllForecastCostResult(ctx context.Context, created PriceInfos) error {

	batchSize := 100
	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		for i := 0; i < len(created); i += batchSize {
			end := i + batchSize
			if end > len(created) {
				end = len(created) // 데이터의 끝을 넘어가지 않도록 조정
			}

			batch := created[i:end]
			if err := d.Create(&batch).Error; err != nil {
				return err
			}
		}
		return nil
	})

	return err

}

func (r *CostRepository) UpsertCostInfo(ctx context.Context, costInfo CostInfo) (int64, int64, error) {
	var updateCount = int64(0)
	var insertCount = int64(0)
	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		err := d.
			Model(costInfo).
			Where(&CostInfo{
				MigrationId:      costInfo.MigrationId,
				Provider:         costInfo.Provider,
				ResourceType:     costInfo.ResourceType,
				Category:         costInfo.Category,
				ActualResourceId: costInfo.ActualResourceId,
				Granularity:      costInfo.Granularity,
				StartDate:        costInfo.StartDate,
				EndDate:          costInfo.EndDate,
			}).First(&costInfo).Error

		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		if err == gorm.ErrRecordNotFound {
			if err := d.Create(&costInfo).Error; err != nil {
				return err
			}
			insertCount++
		} else {
			if err := d.Model(&costInfo).Updates(map[string]interface{}{
				"cost": costInfo.Cost,
				"unit": costInfo.Unit,
			}).Error; err != nil {
				return err
			}

			updateCount++
		}

		return nil
	})

	return updateCount, insertCount, err

}

func (r *CostRepository) GetCostInfoWithFilter(ctx context.Context, param GetCostInfoParam) ([]GetCostInfoResult, error) {
	var costInfo []GetCostInfoResult

	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		query := d.Model(&CostInfo{})

		if len(param.Providers) > 0 {
			query = query.Where("provider IN ?", param.Providers)
		}
		if len(param.ResourceTypes) > 0 {
			query = query.Where("resource_type IN ?", param.ResourceTypes)
		}

		if len(param.ResourceIds) > 0 {
			query = query.Where("actual_resource_id IN ?", param.ResourceIds)
		}

		query = query.Where("start_date >= ? AND end_date <= ?", param.StartDate, param.EndDate)

		switch param.CostAggregationType {
		case constant.Daily:
			query = query.Select("provider, resource_type, category, actual_resource_id, unit, DATE(start_date) AS date, SUM(cost) AS total_cost").
				Group("provider, resource_type, category, actual_resource_id, unit, date")
		case constant.Weekly:
			query = query.Select("provider, resource_type, category, actual_resource_id, unit, DATE_TRUNC('week', start_date) AS date, SUM(cost) AS total_cost").
				Group("provider, resource_type, category, actual_resource_id, unit, date")
		case constant.Monthly:
			query = query.Select("provider, resource_type, category, actual_resource_id, unit, DATE_TRUNC('month', start_date) AS date, SUM(cost) AS total_cost").
				Group("provider, resource_type, category, actual_resource_id, unit, date")
		}

		if param.DateOrder != "" {
			query = query.Order("date " + string(param.DateOrder))
		}

		if param.ResourceTypeOrder != "" {
			query = query.Order("resource_type " + string(param.ResourceTypeOrder))
		}

		if err := query.Find(&costInfo).Error; err != nil {
			return err
		}

		return nil
	})

	return costInfo, err
}
