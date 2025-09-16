package app

import (
	"net/http"
	"strconv"

	"github.com/cloud-barista/cm-ant/internal/core/load"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// createLoadTestScenarioCatalog handler function that creates a new load test scenario catalog.
// @Id CreateLoadTestScenarioCatalog
// @Summary Create Load Test Scenario Catalog
// @Description Create a new load test scenario catalog template with the provided configuration.
// @Tags [Load Test Scenario Catalog Management]
// @Accept json
// @Produce json
// @Param body body load.CreateLoadTestScenarioCatalogReq true "Load Test Scenario Catalog Creation Request"
// @Success 200 {object} app.AntResponse[load.LoadTestScenarioCatalogResult] "Successfully created load test scenario catalog"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters"
// @Failure 500 {object} app.AntResponse[string] "Failed to create load test scenario catalog"
// @Router /api/v1/load/templates/test-scenario-catalogs [post]
func (s *AntServer) createLoadTestScenarioCatalog(c echo.Context) error {
	var req load.CreateLoadTestScenarioCatalogReq
	if err := c.Bind(&req); err != nil {
		log.Error().Err(err).Msg("Failed to bind request")
		return errorResponseJson(http.StatusBadRequest, "Invalid request parameters")
	}

	// Validate required fields
	if req.Name == "" || req.VirtualUsers == "" || req.Duration == "" || req.RampUpTime == "" || req.RampUpSteps == "" {
		return errorResponseJson(http.StatusBadRequest, "Required fields are missing")
	}

	catalog := load.LoadTestScenarioCatalog{
		Name:         req.Name,
		Description:  req.Description,
		VirtualUsers: req.VirtualUsers,
		Duration:     req.Duration,
		RampUpTime:   req.RampUpTime,
		RampUpSteps:  req.RampUpSteps,
	}

	result, err := s.services.loadService.CreateLoadTestScenarioCatalog(c.Request().Context(), catalog)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create load test scenario catalog")
		return errorResponseJson(http.StatusInternalServerError, "Failed to create load test scenario catalog")
	}

	return successResponseJson(c, "Successfully created load test scenario catalog", result)
}

// getAllLoadTestScenarioCatalogs handler function that retrieves all load test scenario catalogs.
// @Id GetAllLoadTestScenarioCatalogs
// @Summary Get All Load Test Scenario Catalogs
// @Description Retrieve a list of all load test scenario catalogs with pagination support.
// @Tags [Load Test Scenario Catalog Management]
// @Accept json
// @Produce json
// @Param page query int false "Page number for pagination (default 1)"
// @Param size query int false "Number of items per page (default 10, max 100)"
// @Param name query string false "Filter by catalog name"
// @Success 200 {object} app.AntResponse[load.GetAllLoadTestScenarioCatalogsResult] "Successfully retrieved load test scenario catalogs"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters"
// @Failure 500 {object} app.AntResponse[string] "Failed to retrieve load test scenario catalogs"
// @Router /api/v1/load/templates/test-scenario-catalogs [get]
func (s *AntServer) getAllLoadTestScenarioCatalogs(c echo.Context) error {
	var req GetAllLoadTestScenarioCatalogsReq
	if err := c.Bind(&req); err != nil {
		log.Error().Err(err).Msg("Failed to bind request")
		return errorResponseJson(http.StatusBadRequest, "Invalid request parameters")
	}

	// Set default values
	if req.Size < 1 || req.Size > 100 {
		req.Size = 10
	}
	if req.Page < 1 {
		req.Page = 1
	}

	param := load.GetAllLoadTestScenarioCatalogsParam{
		Page: req.Page,
		Size: req.Size,
		Name: req.Name,
	}

	result, err := s.services.loadService.GetAllLoadTestScenarioCatalogs(c.Request().Context(), param)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all load test scenario catalogs")
		return errorResponseJson(http.StatusInternalServerError, "Failed to retrieve load test scenario catalogs")
	}

	return successResponseJson(c, "Successfully retrieved load test scenario catalogs", result)
}

// getLoadTestScenarioCatalog handler function that retrieves a specific load test scenario catalog.
// @Id GetLoadTestScenarioCatalog
// @Summary Get Load Test Scenario Catalog
// @Description Retrieve a specific load test scenario catalog by ID.
// @Tags [Load Test Scenario Catalog Management]
// @Accept json
// @Produce json
// @Param id path int true "Load Test Scenario Catalog ID"
// @Success 200 {object} app.AntResponse[load.LoadTestScenarioCatalogResult] "Successfully retrieved load test scenario catalog"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters"
// @Failure 404 {object} app.AntResponse[string] "Load test scenario catalog not found"
// @Failure 500 {object} app.AntResponse[string] "Failed to retrieve load test scenario catalog"
// @Router /api/v1/load/templates/test-scenario-catalogs/{id} [get]
func (s *AntServer) getLoadTestScenarioCatalog(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		log.Error().Err(err).Msg("Invalid ID parameter")
		return errorResponseJson(http.StatusBadRequest, "Invalid ID parameter")
	}

	result, err := s.services.loadService.GetLoadTestScenarioCatalog(c.Request().Context(), uint(id))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get load test scenario catalog")
		if err.Error() == "load test scenario catalog not found" {
			return errorResponseJson(http.StatusNotFound, "Load test scenario catalog not found")
		}
		return errorResponseJson(http.StatusInternalServerError, "Failed to retrieve load test scenario catalog")
	}

	return successResponseJson(c, "Successfully retrieved load test scenario catalog", result)
}

// updateLoadTestScenarioCatalog handler function that updates a load test scenario catalog.
// @Id UpdateLoadTestScenarioCatalog
// @Summary Update Load Test Scenario Catalog
// @Description Update an existing load test scenario catalog with the provided configuration.
// @Tags [Load Test Scenario Catalog Management]
// @Accept json
// @Produce json
// @Param id path int true "Load Test Scenario Catalog ID"
// @Param body body load.UpdateLoadTestScenarioCatalogReq true "Load Test Scenario Catalog Update Request"
// @Success 200 {object} app.AntResponse[load.LoadTestScenarioCatalogResult] "Successfully updated load test scenario catalog"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters"
// @Failure 404 {object} app.AntResponse[string] "Load test scenario catalog not found"
// @Failure 500 {object} app.AntResponse[string] "Failed to update load test scenario catalog"
// @Router /api/v1/load/templates/test-scenario-catalogs/{id} [put]
func (s *AntServer) updateLoadTestScenarioCatalog(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		log.Error().Err(err).Msg("Invalid ID parameter")
		return errorResponseJson(http.StatusBadRequest, "Invalid ID parameter")
	}

	var req load.UpdateLoadTestScenarioCatalogReq
	if err := c.Bind(&req); err != nil {
		log.Error().Err(err).Msg("Failed to bind request")
		return errorResponseJson(http.StatusBadRequest, "Invalid request parameters")
	}

	catalog := load.LoadTestScenarioCatalog{
		Name:         req.Name,
		Description:  req.Description,
		VirtualUsers: req.VirtualUsers,
		Duration:     req.Duration,
		RampUpTime:   req.RampUpTime,
		RampUpSteps:  req.RampUpSteps,
	}

	result, err := s.services.loadService.UpdateLoadTestScenarioCatalog(c.Request().Context(), uint(id), catalog)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update load test scenario catalog")
		if err.Error() == "load test scenario catalog not found" {
			return errorResponseJson(http.StatusNotFound, "Load test scenario catalog not found")
		}
		return errorResponseJson(http.StatusInternalServerError, "Failed to update load test scenario catalog")
	}

	return successResponseJson(c, "Successfully updated load test scenario catalog", result)
}

// deleteLoadTestScenarioCatalog handler function that deletes a load test scenario catalog.
// @Id DeleteLoadTestScenarioCatalog
// @Summary Delete Load Test Scenario Catalog
// @Description Delete a load test scenario catalog by ID.
// @Tags [Load Test Scenario Catalog Management]
// @Accept json
// @Produce json
// @Param id path int true "Load Test Scenario Catalog ID"
// @Success 200 {object} app.AntResponse[string] "Successfully deleted load test scenario catalog"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters"
// @Failure 404 {object} app.AntResponse[string] "Load test scenario catalog not found"
// @Failure 500 {object} app.AntResponse[string] "Failed to delete load test scenario catalog"
// @Router /api/v1/load/templates/test-scenario-catalogs/{id} [delete]
func (s *AntServer) deleteLoadTestScenarioCatalog(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		log.Error().Err(err).Msg("Invalid ID parameter")
		return errorResponseJson(http.StatusBadRequest, "Invalid ID parameter")
	}

	err = s.services.loadService.DeleteLoadTestScenarioCatalog(c.Request().Context(), uint(id))
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete load test scenario catalog")
		if err.Error() == "load test scenario catalog not found" {
			return errorResponseJson(http.StatusNotFound, "Load test scenario catalog not found")
		}
		return errorResponseJson(http.StatusInternalServerError, "Failed to delete load test scenario catalog")
	}

	return successResponseJson(c, "Successfully deleted load test scenario catalog", "Successfully deleted load test scenario catalog")
}
