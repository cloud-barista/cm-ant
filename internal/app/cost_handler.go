package app

import (
	"net/http"
	"strings"

	"github.com/cloud-barista/cm-ant/internal/core/cost"
	"github.com/labstack/echo/v4"
)

func (server *AntServer) getPriceInfo(c echo.Context) error {
	var req PriceInfoReq
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
