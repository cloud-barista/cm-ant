package app

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/cloud-barista/cm-ant/internal/core/cost"
	"github.com/labstack/echo/v4"
)

// @Id UpdatePriceInfos
// @Summar Update Price Information
// @Description Retrieve pricing information for cloud resources based on specified parameters. If saved data is more than 7 days, fetch new data and insert new price data even if same price as before.
// @Tags [Price Management]
// @Accept json
// @Produce json
// @Param body body app.UpdatePriceInfosReq true "Request body containing get price information"
// @Success 200 {object} app.AntResponse[string] "Successfully retrieved pricing information"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters"
// @Failure 500 {object} app.AntResponse[string] "Failed to retrieve pricing information"
// @Router /api/v1/price/info [post]
func (server *AntServer) updatePriceInfos(c echo.Context) error {
	var req UpdatePriceInfosReq
	if err := c.Bind(&req); err != nil {
		return errorResponseJson(http.StatusBadRequest, err.Error())
	}

	if strings.TrimSpace(req.RegionName) == "" ||
		// strings.TrimSpace(req.InstanceType) == "" ||
		strings.TrimSpace(req.ProviderName) == "" {
		return errorResponseJson(http.StatusBadRequest, "provier name, region name, instance type must be set")
	}

	arg := cost.UpdatePriceInfosParam{
		ProviderName: strings.TrimSpace(req.ProviderName),
		RegionName:   strings.TrimSpace(req.RegionName),
		InstanceType: strings.TrimSpace(req.InstanceType),
	}

	err := server.services.costService.UpdatePriceInfos(arg)

	if err != nil {
		return errorResponseJson(http.StatusInternalServerError, err.Error())
	}

	return successResponseJson(
		c,
		"retrieved pricing information",
		fmt.Sprintf("%s csp price information updated. region: %s, instance type: %s", arg.ProviderName, arg.RegionName, arg.InstanceType),
	)
}

// @Id GetPriceInfos
// @Summary Get Price Information
// @Description Retrieve pricing information for cloud resources based on specified query parameters. Returns price data based on provider, region, instance type, vCPU, memory, and OS type. It offer instances with the lowest monthly prices what in the database.
// @Tags Price Management
// @Accept json
// @Produce json
// @Param providerName query string true "Cloud provider name - aws|alibaba|tencent|gcp|azure|ibm"
// @Param regionName query string true "Region name"
// @Param instanceType query string false "Instance type"
// @Param vCpu query string false "Number of vCPUs"
// @Param memory query string false "Amount of memory"
// @Param osType query string false "Operating system type"
// @Success 200 {object} AntResponse[string] "Successfully retrieved pricing information"
// @Failure 400 {object} AntResponse[string] "Invalid request parameters"
// @Failure 500 {object} AntResponse[string] "Failed to retrieve pricing information"
// @Router /api/v1/price/info [get]
func (server *AntServer) getPriceInfos(c echo.Context) error {
	var req GetPriceInfosReq
	if err := c.Bind(&req); err != nil {
		return errorResponseJson(http.StatusBadRequest, err.Error())
	}

	arg := cost.GetPriceInfosParam{
		ProviderName: strings.TrimSpace(req.ProviderName),
		RegionName:   strings.TrimSpace(req.RegionName),
		InstanceType: strings.TrimSpace(req.InstanceType),
		VCpu:         strings.TrimSpace(req.VCpu),
		Memory:       strings.TrimSpace(req.Memory),
		OsType:       strings.TrimSpace(req.OsType),
	}

	r, err := server.services.costService.GetPriceInfos(arg)

	if err != nil {
		return errorResponseJson(http.StatusInternalServerError, err.Error())
	}

	return successResponseJson(
		c,
		"Successfully get price info.",
		r,
	)
}

// @Id UpdateCostInfo
// @Summary Update Cost Information
// @Description Update cost information for specified resources, including details such as migration ID, cost resources, and additional AWS info if applicable. The request body must include a valid migration ID and a list of cost resources. If AWS-specific details are provided, ensure all required fields are populated.
// @Tags [Cost Management]
// @Accept json
// @Produce json
// @Param body body app.UpdateCostInfoReq true "Request body containing cost update information"
// @Success 200 {object} app.AntResponse[cost.UpdateCostInfoResult] "Successfully updated cost information"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters"
// @Failure 500 {object} app.AntResponse[string] "Failed to update cost information"
// @Router /api/v1/cost/info [post]
func (server *AntServer) updateCostInfos(c echo.Context) error {
	var req UpdateCostInfoReq

	if err := c.Bind(&req); err != nil {
		return errorResponseJson(http.StatusBadRequest, "request body binding error")
	}

	if strings.TrimSpace(req.MigrationId) == "" {
		return errorResponseJson(http.StatusBadRequest, "migration id is required")
	}

	if len(req.CostResources) == 0 {
		return errorResponseJson(http.StatusBadRequest, "Migrated resource id list are required")
	}

	costResources := make([]cost.CostResourceParam, 0)

	for _, v := range req.CostResources {
		costResources = append(costResources, cost.CostResourceParam{
			ResourceType: v.ResourceType,
			ResourceIds:  v.ResourceIds,
		})
	}

	endDate := time.Now().Truncate(24*time.Hour).AddDate(0, 0, 1)
	startDate := endDate.AddDate(0, 0, -14)
	param := cost.UpdateCostInfoParam{
		MigrationId:    req.MigrationId,
		Provider:       "aws",
		StartDate:      startDate,
		EndDate:        endDate,
		CostResources:  costResources,
		ConnectionName: req.ConnectionName,
		AwsAdditionalInfo: cost.AwsAdditionalInfoParam{
			OwnerId: req.AwsAdditionalInfo.OwnerId,
			Regions: req.AwsAdditionalInfo.Regions,
		},
	}

	r, err := server.services.costService.UpdateCostInfo(param)

	if err != nil {
		return errorResponseJson(http.StatusInternalServerError, err.Error())
	}

	return successResponseJson(
		c,
		"Successfully updated cost info.",
		r,
	)
}

// @Id GetCostInfo
// @Summary Get Cost Information
// @Description Retrieve cost information for specified parameters within a defined date range. The date range must be within a 6-month period. Optionally, you can specify cost aggregation type and date order for the results.
// @Tags [Cost Management]
// @Accept json
// @Produce json
// @Param startDate query string true "Start date for the cost information retrieval in 'YYYY-MM-DD' format"
// @Param endDate query string true "End date for the cost information retrieval in 'YYYY-MM-DD' format"
// @Param migrationIds query []string false "List of migration IDs to filter the cost information"
// @Param provider query []string false "List of cloud providers to filter the cost information"
// @Param resourceTypes query []string false "List of resource types to filter the cost information"
// @Param resourceIds query []string false "List of resource IDs to filter the cost information"
// @Param costAggregationType query string true "Type of cost aggregation for the results (e.g., 'daily', 'weekly', 'monthly')"
// @Param dateOrder query string false "Order of dates in the result (e.g., 'asc', 'desc')"
// @Param resourceTypeOrder query string false "Order of resource types in the result (e.g., 'asc', 'desc')"
// @Success 200 {object} app.AntResponse[[]cost.GetCostInfoResult] "Successfully retrieved cost information"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters"
// @Failure 500 {object} app.AntResponse[string] "Failed to retrieve cost information"
// @Router /api/v1/cost/info [get]
func (s *AntServer) getCostInfos(c echo.Context) error {
	var req GetCostInfoReq
	if err := c.Bind(&req); err != nil {
		return errorResponseJson(http.StatusBadRequest, "Invalid request parameters")
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)

	if err != nil {
		return errorResponseJson(http.StatusBadRequest, "start date format is incorrect")
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)

	if err != nil {
		return errorResponseJson(http.StatusBadRequest, "end date format is incorrect")
	}

	sixMonthsLater := startDate.AddDate(0, 6, 0)

	if endDate.After(sixMonthsLater) {
		return errorResponseJson(http.StatusBadRequest, "date range must in 6 month")
	}

	if req.CostAggregationType == "" {
		req.CostAggregationType = constant.Daily
	}

	if req.DateOrder == "" {
		req.DateOrder = constant.Asc
	}

	arg := cost.GetCostInfoParam{
		StartDate:           startDate,
		EndDate:             endDate,
		MigrationIds:        req.MigrationIds,
		Providers:           req.Providers,
		ResourceTypes:       req.ResourceTypes,
		ResourceIds:         req.ResourceIds,
		CostAggregationType: req.CostAggregationType,
		DateOrder:           req.DateOrder,
		ResourceTypeOrder:   req.ResourceTypeOrder,
	}

	result, err := s.services.costService.GetCostInfos(arg)

	if err != nil {
		return errorResponseJson(http.StatusInternalServerError, "Failed to retrieve load test result")
	}

	return successResponseJson(c, "Successfully retrieved load test result", result)
}
