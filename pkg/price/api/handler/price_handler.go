package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/cloud-barista/cm-ant/pkg/outbound/spider"
	"github.com/labstack/echo/v4"
)

func Price() echo.HandlerFunc {
	return func(c echo.Context) error {
		regionName := "ap-northeast-2"
		connectionName := "aws-sao-conn"
		productfamily := "ComputeInstance"
		instanceType := "t2.large"
		res, err := spider.GetProductFamilyWitContext(context.Background(), regionName, connectionName)

		if err != nil {
			return err
		}

		if len(res.Productfamily) == 0 {
			return c.JSON(http.StatusOK, map[string]any{
				"message": "bye",
			})
		}

		log.Printf("%+v\n", res)

		priceInfoReq := spider.PriceInfoReq{
			ConnectionName: connectionName,
			FilterList: []spider.FilterReq{
				{
					Key:   "instanceType",
					Value: instanceType,
				},
				// {
				// 	Key:   "vcpu",
				// 	Value: "2",
				// },
				// {
				// 	Key:   "memory",
				// 	Value: "1",
				// },
			},
		}
		r, err := spider.GetPriceInfoWithContext(context.Background(), productfamily, regionName, priceInfoReq)

		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]any{
			"message": "success",
			"result":  r,
		})
	}
}
