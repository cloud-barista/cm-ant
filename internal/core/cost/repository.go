package cost

import (
	"context"
	"fmt"

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

func (r *CostRepository) GetAllMatchingPriceInfoList(ctx context.Context, param GetPriceInfoParam) (PriceInfos, error) {
	var priceInfoList []*PriceInfo

	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		q := d.Model(&PriceInfo{}).
			Where(
				"provider_name = ? AND region_name = ? AND instance_type = ? AND price_policy = ? AND updated_at >= ?",
				param.ProviderName, param.RegionName, param.InstanceType, param.PricePolicy, param.TimeStandard,
			).
			Order("price asc")

		if param.ZoneName != "" {
			q = q.Where("zone_name = ?", param.ZoneName)
		}

		if param.VCpu != "" {
			q = q.Where("v_cpu = ?", param.VCpu)
		}

		if param.Memory != "" {
			q = q.Where("memory = ?", param.Memory)
		}

		if param.Storage != "" {
			q = q.Where("storage = ?", param.Storage)
		}

		if param.OsType != "" {
			q = q.Where("os_type = ?", param.OsType)
		}

		if param.Unit != "" {
			q = q.Where("unit = ?", param.Unit)
		}

		if param.Currency != "" {
			q = q.Where("currency = ?", param.Currency)
		}

		if err := q.Find(&priceInfoList).Error; err != nil {
			return err
		}

		return nil
	})

	return priceInfoList, err

}

func (r *CostRepository) InsertAllResult(ctx context.Context, param GetPriceInfoParam, created PriceInfos) error {

	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		err := d.Create(created).Error

		if err != nil {
			return err
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
