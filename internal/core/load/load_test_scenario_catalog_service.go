package load

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// CreateLoadTestScenarioCatalog creates a new load test scenario catalog.
func (l *LoadService) CreateLoadTestScenarioCatalog(ctx context.Context, catalog LoadTestScenarioCatalog) (LoadTestScenarioCatalogResult, error) {
	log.Info().Msg("Starting CreateLoadTestScenarioCatalog")

	// Check if catalog with same name already exists
	var existingCatalog LoadTestScenarioCatalog
	err := l.db.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", catalog.Name).First(&existingCatalog).Error
	if err == nil {
		return LoadTestScenarioCatalogResult{}, errors.New("catalog with this name already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Msg("Failed to check existing catalog")
		return LoadTestScenarioCatalogResult{}, err
	}

	// Create new catalog
	err = l.db.WithContext(ctx).Create(&catalog).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to create load test scenario catalog")
		return LoadTestScenarioCatalogResult{}, err
	}

	result := LoadTestScenarioCatalogResult{
		ID:           catalog.ID,
		Name:         catalog.Name,
		Description:  catalog.Description,
		VirtualUsers: catalog.VirtualUsers,
		Duration:     catalog.Duration,
		RampUpTime:   catalog.RampUpTime,
		RampUpSteps:  catalog.RampUpSteps,
		CreatedAt:    catalog.CreatedAt,
		UpdatedAt:    catalog.UpdatedAt,
	}

	log.Info().Uint("catalogId", catalog.ID).Msg("Successfully created load test scenario catalog")
	return result, nil
}

// GetAllLoadTestScenarioCatalogs retrieves all load test scenario catalogs with pagination.
func (l *LoadService) GetAllLoadTestScenarioCatalogs(ctx context.Context, param GetAllLoadTestScenarioCatalogsParam) (GetAllLoadTestScenarioCatalogsResult, error) {
	log.Info().Msg("Starting GetAllLoadTestScenarioCatalogs")

	var catalogs []LoadTestScenarioCatalog
	var totalCount int64

	query := l.db.WithContext(ctx).Model(&LoadTestScenarioCatalog{}).Where("deleted_at IS NULL")

	// Apply name filter if provided
	if param.Name != "" {
		query = query.Where("name LIKE ?", "%"+param.Name+"%")
	}

	// Get total count
	err := query.Count(&totalCount).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to count load test scenario catalogs")
		return GetAllLoadTestScenarioCatalogsResult{}, err
	}

	// Apply pagination and get results
	offset := (param.Page - 1) * param.Size
	err = query.Offset(offset).Limit(param.Size).Order("created_at DESC").Find(&catalogs).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to get load test scenario catalogs")
		return GetAllLoadTestScenarioCatalogsResult{}, err
	}

	// Convert to result format
	var catalogResults []LoadTestScenarioCatalogResult
	for _, catalog := range catalogs {
		catalogResults = append(catalogResults, LoadTestScenarioCatalogResult{
			ID:           catalog.ID,
			Name:         catalog.Name,
			Description:  catalog.Description,
			VirtualUsers: catalog.VirtualUsers,
			Duration:     catalog.Duration,
			RampUpTime:   catalog.RampUpTime,
			RampUpSteps:  catalog.RampUpSteps,
			CreatedAt:    catalog.CreatedAt,
			UpdatedAt:    catalog.UpdatedAt,
		})
	}

	result := GetAllLoadTestScenarioCatalogsResult{
		LoadTestScenarioCatalogs: catalogResults,
		TotalRow:                 totalCount,
	}

	log.Info().Int64("totalCount", totalCount).Int("returnedCount", len(catalogResults)).Msg("Successfully retrieved load test scenario catalogs")
	return result, nil
}

// GetLoadTestScenarioCatalog retrieves a specific load test scenario catalog by ID.
func (l *LoadService) GetLoadTestScenarioCatalog(ctx context.Context, id uint) (LoadTestScenarioCatalogResult, error) {
	log.Info().Uint("catalogId", id).Msg("Starting GetLoadTestScenarioCatalog")

	var catalog LoadTestScenarioCatalog
	err := l.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&catalog).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return LoadTestScenarioCatalogResult{}, errors.New("load test scenario catalog not found")
		}
		log.Error().Err(err).Msg("Failed to get load test scenario catalog")
		return LoadTestScenarioCatalogResult{}, err
	}

	result := LoadTestScenarioCatalogResult{
		ID:           catalog.ID,
		Name:         catalog.Name,
		Description:  catalog.Description,
		VirtualUsers: catalog.VirtualUsers,
		Duration:     catalog.Duration,
		RampUpTime:   catalog.RampUpTime,
		RampUpSteps:  catalog.RampUpSteps,
		CreatedAt:    catalog.CreatedAt,
		UpdatedAt:    catalog.UpdatedAt,
	}

	log.Info().Uint("catalogId", id).Msg("Successfully retrieved load test scenario catalog")
	return result, nil
}

// UpdateLoadTestScenarioCatalog updates an existing load test scenario catalog.
func (l *LoadService) UpdateLoadTestScenarioCatalog(ctx context.Context, id uint, catalog LoadTestScenarioCatalog) (LoadTestScenarioCatalogResult, error) {
	log.Info().Uint("catalogId", id).Msg("Starting UpdateLoadTestScenarioCatalog")

	// Check if catalog exists
	var existingCatalog LoadTestScenarioCatalog
	err := l.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&existingCatalog).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return LoadTestScenarioCatalogResult{}, errors.New("load test scenario catalog not found")
		}
		log.Error().Err(err).Msg("Failed to check existing catalog")
		return LoadTestScenarioCatalogResult{}, err
	}

	// Check if name is being changed and if new name already exists
	if catalog.Name != "" && catalog.Name != existingCatalog.Name {
		var nameCheckCatalog LoadTestScenarioCatalog
		err = l.db.WithContext(ctx).Where("name = ? AND id != ? AND deleted_at IS NULL", catalog.Name, id).First(&nameCheckCatalog).Error
		if err == nil {
			return LoadTestScenarioCatalogResult{}, errors.New("catalog with this name already exists")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Msg("Failed to check existing catalog name")
			return LoadTestScenarioCatalogResult{}, err
		}
	}

	// Update only provided fields
	updateData := make(map[string]interface{})
	if catalog.Name != "" {
		updateData["name"] = catalog.Name
	}
	if catalog.Description != "" {
		updateData["description"] = catalog.Description
	}
	if catalog.VirtualUsers != "" {
		updateData["virtual_users"] = catalog.VirtualUsers
	}
	if catalog.Duration != "" {
		updateData["duration"] = catalog.Duration
	}
	if catalog.RampUpTime != "" {
		updateData["ramp_up_time"] = catalog.RampUpTime
	}
	if catalog.RampUpSteps != "" {
		updateData["ramp_up_steps"] = catalog.RampUpSteps
	}

	err = l.db.WithContext(ctx).Model(&existingCatalog).Updates(updateData).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to update load test scenario catalog")
		return LoadTestScenarioCatalogResult{}, err
	}

	// Get updated catalog
	err = l.db.WithContext(ctx).Where("id = ?", id).First(&existingCatalog).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to get updated catalog")
		return LoadTestScenarioCatalogResult{}, err
	}

	result := LoadTestScenarioCatalogResult{
		ID:           existingCatalog.ID,
		Name:         existingCatalog.Name,
		Description:  existingCatalog.Description,
		VirtualUsers: existingCatalog.VirtualUsers,
		Duration:     existingCatalog.Duration,
		RampUpTime:   existingCatalog.RampUpTime,
		RampUpSteps:  existingCatalog.RampUpSteps,
		CreatedAt:    existingCatalog.CreatedAt,
		UpdatedAt:    existingCatalog.UpdatedAt,
	}

	log.Info().Uint("catalogId", id).Msg("Successfully updated load test scenario catalog")
	return result, nil
}

// DeleteLoadTestScenarioCatalog soft deletes a load test scenario catalog.
func (l *LoadService) DeleteLoadTestScenarioCatalog(ctx context.Context, id uint) error {
	log.Info().Uint("catalogId", id).Msg("Starting DeleteLoadTestScenarioCatalog")

	// Check if catalog exists
	var catalog LoadTestScenarioCatalog
	err := l.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&catalog).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("load test scenario catalog not found")
		}
		log.Error().Err(err).Msg("Failed to check existing catalog")
		return err
	}

	// Soft delete
	err = l.db.WithContext(ctx).Delete(&catalog).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete load test scenario catalog")
		return err
	}

	log.Info().Uint("catalogId", id).Msg("Successfully deleted load test scenario catalog")
	return nil
}
