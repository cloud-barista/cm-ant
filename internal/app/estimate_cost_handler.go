package app

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cloud-barista/cm-ant/internal/config"
	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/cloud-barista/cm-ant/internal/core/cost"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// @Id UpdateAndGetEstimateCost
// @Summary Update and Retrieve Estimated Cost Information
// @Description Update the estimate cost based on provided specifications and retrieve the updated cost estimation. Required fields for each specification include `ProviderName`, `RegionName`, and `InstanceType`. Specifications can also be provided in a formatted string using `+` delimiter.
// @Tags [Cost Estimate]
// @Accept json
// @Produce json
// @Param body body UpdateAndGetEstimateCostReq true "Request body for updating and retrieving estimated cost information"
// @Success 200 {object} app.AntResponse[cost.EstimateCostResults] "Successfully updated and retrieved estimated cost information"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters or format"
// @Failure 500 {object} app.AntResponse[string] "Failed to update or retrieve estimated cost information"
// @Router /api/v1/cost/estimate [post]
func (a *AntServer) updateAndGetEstimateCost(c echo.Context) error {
	var req UpdateAndGetEstimateCostReq
	if err := c.Bind(&req); err != nil {
		return errorResponseJson(http.StatusBadRequest, err.Error())
	}

	if len(req.Specs) == 0 && len(req.SpecsWithFormat) == 0 {
		return errorResponseJson(http.StatusBadRequest, "request is invalid. check the required request body properties")
	}

	pastTime := time.Now().Add(-config.AppConfig.Cost.Estimation.UpdateInterval)

	recommendSpecs := make([]cost.RecommendSpecParam, 0)

	if len(req.Specs) > 0 {
		for _, v := range req.Specs {
			pn := strings.TrimSpace(strings.ToLower(v.ProviderName))
			rn := strings.TrimSpace(v.RegionName)
			it := strings.TrimSpace(v.InstanceType)

			if pn == "" || rn == "" || it == "" {
				return errorResponseJson(http.StatusBadRequest, "request is invalid. check the required request body properties")
			}

			param := cost.RecommendSpecParam{
				ProviderName: pn,
				RegionName:   rn,
				InstanceType: it,
				Image:        strings.TrimSpace(v.Image),
			}
			recommendSpecs = append(recommendSpecs, param)
		}
	}

	if len(req.SpecsWithFormat) > 0 {
		delim := "+"

		for _, v := range req.SpecsWithFormat {

			cs := strings.TrimSpace(v.CommonSpec)
			ci := strings.TrimSpace(v.CommonImage)

			if cs == "" {
				return errorResponseJson(http.StatusBadRequest, "request is invalid. check the required request body properties")
			}

			splitedCommonSpec := strings.Split(cs, delim)
			splitedCommonImage := strings.Split(ci, delim)

			if len(splitedCommonSpec) != 3 {
				log.Error().Msgf("common spec format is not correct; image: %s; spec: %s", ci, cs)
				return errorResponseJson(http.StatusBadRequest, fmt.Sprintf("common spec format is not correct; image: %s; spec: %s", ci, cs))
			}

			if len(splitedCommonImage) == 3 && (splitedCommonImage[0] != splitedCommonSpec[0] || splitedCommonImage[1] != splitedCommonSpec[1]) {
				log.Error().Msgf("common image and spec recommendation is wrong; image: %s; spec: %s", ci, cs)
				return errorResponseJson(http.StatusBadRequest, fmt.Sprintf("common image and spec recommendation is wrong; image: %s; spec: %s", ci, cs))
			}

			pn := strings.TrimSpace(strings.ToLower(splitedCommonSpec[0]))
			rn := strings.TrimSpace(splitedCommonSpec[1])
			it := strings.TrimSpace(splitedCommonSpec[2])

			if pn == "" || rn == "" || it == "" {
				return errorResponseJson(http.StatusBadRequest, "request is invalid. check the required request body properties")
			}

			param := cost.RecommendSpecParam{
				ProviderName: pn,
				RegionName:   rn,
				InstanceType: it,
			}

			if len(splitedCommonImage) == 3 {
				param.Image = strings.TrimSpace(splitedCommonImage[2])
			}

			recommendSpecs = append(recommendSpecs, param)
		}
	}

	arg := cost.UpdateAndGetEstimateCostParam{
		RecommendSpecs: recommendSpecs,
		TimeStandard:   time.Date(pastTime.Year(), pastTime.Month(), pastTime.Day(), 0, 0, 0, 0, pastTime.Location()),
		PricePolicy:    constant.OnDemand,
	}

	res, err := a.services.costService.UpdateAndGetEstimateCost(arg)

	if err != nil {
		return errorResponseJson(http.StatusInternalServerError, err.Error())
	}

	return successResponseJson(
		c,
		"Successfully update and get estimate cost info",
		res,
	)
}

// @Id UpdateEstimateForecastCost
// @Summary Update and Retrieve Estimated Forecast Cost
// @Description Update and retrieve forecasted cost estimates for a specified namespace and migration configuration ID over the past 14 days.
// @Tags [Cost Estimate]
// @Accept json
// @Produce json
// @Param body body UpdateEstimateForecastCostReq true "Request body containing NsId (Namespace ID) and MciId (Migration Configuration ID) for cost estimation forecast"
// @Success 200 {object} app.AntResponse[cost.UpdateEstimateForecastCostInfoResult] "Successfully updated and retrieved estimated forecast cost information"
// @Failure 400 {object} app.AntResponse[string] "Request body binding error"
// @Failure 500 {object} app.AntResponse[string] "Failed to update or retrieve forecast cost information"
// @Router /api/v1/cost/estimate/forecast [post]
func (a *AntServer) updateEstimateForecastCost(c echo.Context) error {
	var req UpdateEstimateForecastCostReq

	if err := c.Bind(&req); err != nil {
		return errorResponseJson(http.StatusBadRequest, "request body binding error")
	}

	endDate := time.Now().Truncate(24*time.Hour).AddDate(0, 0, 1)
	startDate := endDate.AddDate(0, 0, -14)

	param := cost.UpdateEstimateForecastCostParam{
		NsId:      req.NsId,
		MciId:     req.MciId,
		StartDate: startDate,
		EndDate:   endDate,
	}

	r, err := a.services.costService.UpdateEstimateForecastCost(param)

	if err != nil {
		return errorResponseJson(http.StatusInternalServerError, err.Error())
	}

	return successResponseJson(
		c,
		"Successfully update estimate forecast cost info.",
		r,
	)
}

// @Id GetEstimateCost
// @Summary Retrieve Estimated Cost Information
// @Description Fetch estimated cost details based on provider, region, instance type, and resource specifications. Pagination support is provided through `Page` and `Size` parameters.
// @Tags [Cost Estimate]
// @Accept json
// @Produce json
// @Param providerName query string false "Cloud provider name to filter estimated costs"
// @Param regionName query string false "Region name to filter estimated costs"
// @Param instanceType query string false "Instance type to filter estimated costs"
// @Param vCpu query string false "Number of vCPUs to filter estimated costs"
// @Param memory query string false "Memory size to filter estimated costs"
// @Param osType query string false "Operating system type to filter estimated costs"
// @Param page query int false "Page number for pagination (default: 1)"
// @Param size query int false "Number of records per page (default: 100, max: 100)"
// @Success 200 {object} app.AntResponse[cost.EstimateCostInfoResults] "Successfully retrieved estimated cost information"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters"
// @Failure 500 {object} app.AntResponse[string] "Failed to retrieve estimated cost information"
// @Router /api/v1/cost/estimate [get]
func (server *AntServer) getEstimateCost(c echo.Context) error {
	var req GetEstimateCostInfosReq
	if err := c.Bind(&req); err != nil {
		return errorResponseJson(http.StatusBadRequest, err.Error())
	}

	if req.Page < 1 {
		req.Page = 1
	}

	if req.Size < 1 || req.Size > 100 {
		req.Size = 100
	}

	arg := cost.GetEstimateCostParam{
		ProviderName: strings.TrimSpace(req.ProviderName),
		RegionName:   strings.TrimSpace(req.RegionName),
		InstanceType: strings.TrimSpace(req.InstanceType),
		VCpu:         strings.TrimSpace(req.VCpu),
		Memory:       strings.TrimSpace(req.Memory),
		OsType:       strings.TrimSpace(req.OsType),
		Page:         req.Page,
		Size:         req.Size,
	}

	r, err := server.services.costService.GetEstimateCost(arg)

	if err != nil {
		return errorResponseJson(http.StatusInternalServerError, err.Error())
	}

	return successResponseJson(
		c,
		"Successfully get price info.",
		r,
	)
}

// @Id GetEstimateForecastCost
// @Summary Retrieve Estimated Forecast Cost Information
// @Description Fetch estimated forecast cost data based on specified parameters, including a date range that must be within 6 months. Supports pagination and filtering by namespace IDs, migration configuration IDs, and resource types.
// @Tags [Cost Estimate]
// @Accept json
// @Produce json
// @Param startDate query string true "Start date for the forecast cost retrieval in 'YYYY-MM-DD' format"
// @Param endDate query string true "End date for the forecast cost retrieval in 'YYYY-MM-DD' format"
// @Param nsIds query []string false "List of namespace IDs to filter forecast cost information"
// @Param mciIds query []string false "List of migration configuration IDs to filter forecast cost information"
// @Param providers query []string false "List of cloud providers to filter forecast cost information"
// @Param resourceTypes query []string false "List of resource types to filter forecast cost information"
// @Param resourceIds query []string false "List of resource IDs to filter forecast cost information"
// @Param costAggregationType query string false "Type of cost aggregation (e.g., 'daily', 'weekly', 'monthly')"
// @Param dateOrder query string false "Order of dates in the result (e.g., 'asc', 'desc')"
// @Param resourceTypeOrder query string false "Order of resource types in the result (e.g., 'asc', 'desc')"
// @Param page query int false "Page number for pagination (default: 1)"
// @Param size query int false "Number of records per page (default: 10000, max: 10000)"
// @Success 200 {object} app.AntResponse[cost.GetEstimateForecastCostInfoResults] "Successfully retrieved estimated forecast cost information"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters or date format errors"
// @Failure 500 {object} app.AntResponse[string] "Failed to retrieve estimated forecast cost information"
// @Router /api/v1/cost/estimate/forecast [get]
func (s *AntServer) getEstimateForecastCost(c echo.Context) error {
	var req GetEstimateForecastCostReq
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

	if endDate.Before(startDate) {
		return errorResponseJson(http.StatusBadRequest, "end date must be after than start date")
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

	if req.Page < 1 {
		req.Page = 1
	}

	if req.Size < 1 || req.Size > 10000 {
		req.Size = 10000
	}

	arg := cost.GetEstimateForecastCostParam{
		Page:                req.Page,
		Size:                req.Size,
		StartDate:           startDate,
		EndDate:             endDate,
		NsIds:               req.NsIds,
		MciIds:              req.MciIds,
		Providers:           req.Providers,
		ResourceTypes:       req.ResourceTypes,
		ResourceIds:         req.ResourceIds,
		CostAggregationType: req.CostAggregationType,
		DateOrder:           req.DateOrder,
		ResourceTypeOrder:   req.ResourceTypeOrder,
	}

	result, err := s.services.costService.GetEstimateForecastCostInfos(arg)

	if err != nil {
		return errorResponseJson(http.StatusInternalServerError, "Failed to get estimate forecast cost")
	}

	return successResponseJson(c, "Successfully get estimate forecast cost", result)
}

// @Id UpdateEstimateForecastCostRaw
// @Summary Update and Retrieve Raw Estimated Forecast Cost
// @Description Update and retrieve raw forecasted cost estimates for specified cost resources and additional AWS information over the past 14 days.
// @Tags [Cost Estimate]
// @Accept json
// @Produce json
// @Param body body UpdateEstimateForecastCostRawReq true "Request body containing details for cost estimation forecast"
// @Success 200 {object} app.AntResponse[cost.UpdateEstimateForecastCostInfoResult] "Successfully updated and retrieved raw estimated forecast cost information in raw data"
// @Failure 400 {object} app.AntResponse[string] "Migrated resource id list is required"
// @Failure 500 {object} app.AntResponse[string] "Error updating or retrieving forecast cost information"
// @Router /api/v1/cost/estimate/forecast/raw [post]
func (server *AntServer) updateEstimateForecastCostRaw(c echo.Context) error {
	var req UpdateEstimateForecastCostRawReq

	if err := c.Bind(&req); err != nil {
		return errorResponseJson(http.StatusBadRequest, "request body binding error")
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
	param := cost.UpdateEstimateForecastCostRawParam{
		Provider:      "aws",
		StartDate:     startDate,
		EndDate:       endDate,
		CostResources: costResources,
		AwsAdditionalInfo: cost.AwsAdditionalInfoParam{
			OwnerId: req.AwsAdditionalInfo.OwnerId,
			Regions: req.AwsAdditionalInfo.Regions,
		},
	}

	r, err := server.services.costService.UpdateEstimateForecastCostRaw(param)

	if err != nil {
		return errorResponseJson(http.StatusInternalServerError, err.Error())
	}

	return successResponseJson(
		c,
		"Successfully updated cost info.",
		r,
	)
}
