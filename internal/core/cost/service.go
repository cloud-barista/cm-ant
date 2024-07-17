package cost

import (
	"context"
	"time"

	"github.com/cloud-barista/cm-ant/internal/infra/outbound/spider"
)

type CostService struct {
	costRepo     *CostRepository
	spiderClient *spider.SpiderClient
}

func NewCostService(costRepo *CostRepository, client *spider.SpiderClient) *CostService {
	return &CostService{
		costRepo:     costRepo,
		spiderClient: client,
	}
}

type PriceInfoParam struct {
	RegionName     string        `json:"regionName"`
	ConnectionName string        `json:"ConnectionName"`
	FilterList     []FilterParam `json:"FilterList"`
}

type FilterParam struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

func (c *CostService) GetPriceInfo(param PriceInfoParam) (spider.CloudPriceDataRes, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var fl []spider.FilterReq

	for _, v := range param.FilterList {
		fl = append(fl, spider.FilterReq{
			Key:   v.Key,
			Value: v.Value,
		})
	}
	req := spider.PriceInfoReq{
		ConnectionName: param.ConnectionName,
		FilterList:     fl,
	}
	result, err := c.spiderClient.GetPriceInfoWithContext(ctx, "ComputeInstance", param.RegionName, req)
	if err != nil {
		return spider.CloudPriceDataRes{}, err
	}

	return result, nil
}
