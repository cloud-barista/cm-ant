package app

import (
	"net/http"
	"strings"

	"github.com/cloud-barista/cm-ant/internal/core/cost"
	"github.com/labstack/echo/v4"
)

var onDemandPricingPolicyMap = map[string]string{
	"aws":        "OnDemand",
	"gcp":        "OnDemand",
	"azure":      "",
	"tencent":    "POSTPAID_BY_HOUR",
	"alibaba":    "PayAsYouGo",
	"ibm":        "",
	"ncp":        "",
	"ncpvpc":     "",
	"nhncloud":   "",
	"ktcloud":    "",
	"ktcloudvpc": "",
}

func (server *AntServer) getPriceInfo(c echo.Context) error {
	var req PriceInfoReq
	if err := c.Bind(&req); err != nil {
		return err
	}

	arg := GetPriceInfoParam(req)

	r, err := server.services.costService.GetPriceInfo(arg)

	if err != nil {
		return errorResponseJson(http.StatusBadRequest, err.Error())
	}

	return successResponseJson(
		c,
		"Successfully get price info.",
		r,
	)
}

func GetPriceInfoParam(req PriceInfoReq) cost.PriceInfoParam {

	providerName := strings.ToLower(req.ProviderName)

	param := cost.PriceInfoParam{
		RegionName:     req.RegionName,
		ConnectionName: req.ConnectionName,
		FilterList: []cost.FilterParam{
			{
				Key:   "instanceType",
				Value: req.CspSpecName,
			},
			{
				Key:   "pricingPolicy",
				Value: onDemandPricingPolicyMap[providerName],
			},
		},
	}

	if req.VCpu != "" {
		param.FilterList = append(param.FilterList, cost.FilterParam{
			Key:   "vcpu",
			Value: req.VCpu,
		})
	} else if req.MemoryGiB != "" {
		param.FilterList = append(param.FilterList, cost.FilterParam{
			Key:   "memory",
			Value: req.MemoryGiB,
		})
	} else if req.OsType != "" {
		param.FilterList = append(param.FilterList, cost.FilterParam{
			Key:   "operatingSystem",
			Value: req.OsType,
		})
	}

	return param
}
