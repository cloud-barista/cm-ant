package app

import (
	"net/http"
	"strings"
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/cost"
	"github.com/labstack/echo/v4"
)

// getPriceInfo handler function that retrieves pricing information for cloud resources.
// @Id GetPriceInfo
// @Summary Get Price Information
// @Description Retrieve pricing information for cloud resources based on specified parameters.
// @Tags [Pricing Management]
// @Accept json
// @Produce json
// @Param RegionName query string true "Name of the region"
// @Param ConnectionName query string true "Name of the connection"
// @Param ProviderName query string true "Name of the cloud provider"
// @Param InstanceType query string true "Type of the instance"
// @Param ZoneName query string false "Name of the zone"
// @Param VCpu query string false "Number of virtual CPUs"
// @Param Memory query string false "Amount of memory. Don't need to pass unit like 'gb'"
// @Param Storage query string false "Amount of storage"
// @Param OsType query string false "Operating system type"
// @Success 200 {object} app.AntResponse[cost.AllPriceInfoResult] "Successfully retrieved pricing information"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters"
// @Failure 500 {object} app.AntResponse[string] "Failed to retrieve pricing information"
// @Router /api/v1/price/info [get]
func (server *AntServer) getPriceInfo(c echo.Context) error {
	var req GetPriceInfoReq
	if err := c.Bind(&req); err != nil {
		return errorResponseJson(http.StatusBadRequest, err.Error())
	}

	if strings.TrimSpace(req.RegionName) == "" ||
		strings.TrimSpace(req.ConnectionName) == "" ||
		strings.TrimSpace(req.ProviderName) == "" ||
		strings.TrimSpace(req.InstanceType) == "" {
		return errorResponseJson(http.StatusBadRequest, "Region Name, Connection Name, Provider Name, Instance type must be set")
	}

	arg := cost.GetPriceInfoParam{
		ProviderName:   strings.TrimSpace(req.ProviderName),
		ConnectionName: strings.TrimSpace(req.ConnectionName),
		RegionName:     strings.TrimSpace(req.RegionName),
		InstanceType:   strings.TrimSpace(req.InstanceType),
		ZoneName:       strings.TrimSpace(req.ZoneName),
		VCpu:           strings.TrimSpace(req.VCpu),
		Memory:         strings.TrimSpace(req.Memory),
		Storage:        strings.TrimSpace(req.Storage),
		OsType:         strings.TrimSpace(req.OsType),
	}

	r, err := server.services.costService.GetPriceInfo(arg)

	if err != nil {
		return errorResponseJson(http.StatusInternalServerError, err.Error())
	}

	return successResponseJson(
		c,
		"Successfully get price info.",
		r,
	)
}

func (server *AntServer) updateCostInfo(c echo.Context) error {
	var req UpdateCostInfo

	if err := c.Bind(&req); err != nil {
		return errorResponseJson(http.StatusBadRequest, "request body binding error")
	}

	if len(req.CostResources) == 0 {
		return errorResponseJson(http.StatusBadRequest, "Migrated resource id list should pass")
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
		ConnectionName: req.ConnectionName,
		StartDate:      startDate,
		EndDate:        endDate,
		CostResources:  costResources,
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
		"Successfully get cost info.",
		r,
	)
}
