package handler

import (
	"context"
	"log"
	"net/http"
	"strings"

	spider "github.com/cloud-barista/cm-ant/pkg/outbound/spider"
	api "github.com/cloud-barista/cm-ant/pkg/price/api"
	"github.com/labstack/echo/v4"
)

var productFamilyMap = map[string]string{
	"aws":     "Compute Instance",
	"gcp":     "Compute",
	"tencent": "cvm",
	"alibaba": "ecs",
}

var onDemandPricingPolicyMap = map[string]string{
	"aws":     "OnDemand",
	"gcp":     "OnDemand",
	"tencent": "POSTPAID_BY_HOUR",
	"alibaba": "PayAsYouGo",
}

// var reservedPricingPolicyMap = map[string]string{
// 	"aws":     "Reserved",
// 	"gcp":     "Commit1Yr",
// 	"tencent": "RESERVED",  // PREPAID, SPOTPAID
// 	"alibaba": "Subscription",
// }

func GetPriceInfoHandler() echo.HandlerFunc {
	return func(c echo.Context) error {

		var priceInfoReq api.PriceInfoReq
		if err := c.Bind(&priceInfoReq); err != nil {
			return err
		}

		providerName := strings.ToLower(priceInfoReq.ProviderName)
		priceInfoReq.ProviderName = providerName

		regionName := priceInfoReq.RegionName
		connectionName := priceInfoReq.ConnectionName
		res, err := spider.GetProductFamilyWitContext(context.Background(), regionName, connectionName)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{
				"status":  "internal server error",
				"message": err.Error(),
			})
		}

		if len(res.Productfamily) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"message": "doesn't exist product family on csp : " + priceInfoReq.ProviderName,
			})
		}
		log.Printf("%+v\n", res)

		var productFamily string
		for _, p := range res.Productfamily {
			if strings.EqualFold(p, productFamilyMap[providerName]) {
				productFamily = p
			}
		}

		if productFamily == "" {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"message": "can't find matching product family on csp : " + priceInfoReq.ProviderName,
			})
		}

		getPriceInfoReq := getPriceInfoRequest(priceInfoReq)

		r, err := spider.GetPriceInfoWithContext(context.Background(), productFamily, regionName, getPriceInfoReq)

		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]any{
			"message": "success",
			"result":  r,
		})
	}
}

func getPriceInfoRequest(priceInfoReq api.PriceInfoReq) spider.PriceInfoReq {
	getPriceInfoReq := spider.PriceInfoReq{
		ConnectionName: priceInfoReq.ConnectionName,
		FilterList: []spider.FilterReq{
			{
				Key:   "instanceType",
				Value: priceInfoReq.CspSpecName,
			},
			{
				Key:   "pricingPolicy",
				Value: onDemandPricingPolicyMap[priceInfoReq.ProviderName],
			},
		},
	}

	if priceInfoReq.VCpu != "" {
		getPriceInfoReq.FilterList = append(getPriceInfoReq.FilterList, spider.FilterReq{
			Key:   "vcpu",
			Value: priceInfoReq.VCpu,
		})
	} else if priceInfoReq.MemoryGiB != "" {
		getPriceInfoReq.FilterList = append(getPriceInfoReq.FilterList, spider.FilterReq{
			Key:   "memory",
			Value: priceInfoReq.MemoryGiB,
		})
	} else if priceInfoReq.OsType != "" {
		getPriceInfoReq.FilterList = append(getPriceInfoReq.FilterList, spider.FilterReq{
			Key:   "operatingSystem",
			Value: priceInfoReq.OsType,
		})
	}

	return getPriceInfoReq
}
