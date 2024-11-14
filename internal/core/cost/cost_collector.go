package cost

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/cloud-barista/cm-ant/internal/infra/outbound/spider"
	"github.com/cloud-barista/cm-ant/internal/infra/outbound/tumblebug"
	"github.com/cloud-barista/cm-ant/internal/utils"
	"github.com/rs/zerolog/log"
)

type CostCollector interface {
	Readyz(context.Context) error
	UpdateEstimateForecastCost(context.Context, UpdateEstimateForecastCostParam) (EstimateForecastCostInfos, error)
	GetCostInfos(context.Context, UpdateEstimateForecastCostRawParam) (EstimateForecastCostInfos, error)
}

type AwsCostExplorerBaristaCostCollector struct {
	sc *spider.SpiderClient
	tc *tumblebug.TumblebugClient
}

func NewAwsCostExplorerSpiderCostCollector(sc *spider.SpiderClient, tc *tumblebug.TumblebugClient) CostCollector {
	return &AwsCostExplorerBaristaCostCollector{
		sc: sc,
		tc: tc,
	}
}

func (a *AwsCostExplorerBaristaCostCollector) Readyz(ctx context.Context) error {
	err := a.sc.ReadyzWithContext(ctx)
	if err != nil {
		return err
	}

	return nil
}

var (
	resourceFilterMap = map[constant.ResourceType][]constant.AwsService{
		constant.VM: {
			constant.AwsEC2,
			constant.AwsEC2Other,
		},
		constant.VNet: {
			constant.AwsVpc,
		},
		constant.DataDisk: {
			constant.AwsEC2Other,
		},
	}
	serviceToResourceType = map[constant.AwsService]constant.ResourceType{
		constant.AwsEC2:          constant.VM,
		constant.AwsEC2Other:     constant.VM,
		constant.AwsVpc:          constant.VNet,
		constant.AwsCostExplorer: constant.Etc,
		constant.AwsTax:          constant.Etc,
	}
)

type costWithResourceReq struct {
	StartDate   string           `json:"startDate"`
	EndDate     string           `json:"endDate"`
	Granularity string           `json:"granularity"` // HOURLY, DAILY, MONTHLY
	Metrics     []string         `json:"metrics"`
	Filter      filterExpression `json:"filter"`
	Groups      []groupBy        `json:"groups"`
}

type filterExpression struct {
	And            []*filterExpression `json:"and,omitempty"`
	Or             []*filterExpression `json:"or,omitempty"`
	Not            *filterExpression   `json:"not,omitempty"`
	CostCategories *keyValues          `json:"costCategories,omitempty"`
	Dimensions     *keyValues          `json:"dimensions,omitempty"`
	Tags           *keyValues          `json:"tags,omitempty"`
}

type keyValues struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

type groupBy struct {
	Key  string `json:"key"`
	Type string `json:"type"` // DIMENSION | TAG | COST_CATEGORY
}

func (a *AwsCostExplorerBaristaCostCollector) generateFilterValue(
	costResources []CostResourceParam, awsAdditionalInfo AwsAdditionalInfoParam,
) (
	[]string, []string, error,
) {
	var serviceValue = make([]string, 0)
	var resourceIdValues = make([]string, 0)

	for _, cr := range costResources {
		resources, ok := resourceFilterMap[cr.ResourceType]
		if !ok {
			continue
		}

		for _, n := range resources {
			serviceValue = append(serviceValue, string(n))
		}

		if cr.ResourceType == constant.VNet {
			var idsTmp = make([]string, 0)
			for _, r := range awsAdditionalInfo.Regions {
				for _, id := range cr.ResourceIds {
					resourceId := fmt.Sprintf("arn:aws:ec2:%s:%s:network-interface/%s", r, awsAdditionalInfo.OwnerId, id)
					idsTmp = append(idsTmp, resourceId)
				}
			}

			resourceIdValues = append(resourceIdValues, idsTmp...)
		} else {
			resourceIdValues = append(resourceIdValues, cr.ResourceIds...)
		}
	}

	return serviceValue, resourceIdValues, nil
}

func (a *AwsCostExplorerBaristaCostCollector) GetCostInfos(ctx context.Context, param UpdateEstimateForecastCostRawParam) (EstimateForecastCostInfos, error) {

	if param.ConnectionName == "" {
		param.ConnectionName = costExplorerConnectionName
	}

	serviceFilterValue, resourceIdFilterValue, err := a.generateFilterValue(param.CostResources, param.AwsAdditionalInfo)
	if err != nil {
		log.Error().Msgf("parsing service and resource id for filtering cost explorer value")
		return nil, err
	}

	if len(serviceFilterValue) == 0 || len(resourceIdFilterValue) == 0 {
		return nil, ErrRequestResourceEmpty
	}

	granularity := "DAILY"
	metrics := "UnblendedCost"
	costWithResourceReq := costWithResourceReq{
		StartDate:   param.StartDate.Format("2006-01-02"),
		EndDate:     param.EndDate.Format("2006-01-02"),
		Granularity: granularity,
		Metrics:     []string{metrics}, // BlendedCost, UnblendedCost, ...
		Filter: filterExpression{
			Or: []*filterExpression{
				{
					And: []*filterExpression{
						{
							Dimensions: &keyValues{
								Key:    "RESOURCE_ID",
								Values: resourceIdFilterValue,
							},
						},
						{
							Dimensions: &keyValues{
								Key:    "SERVICE",
								Values: serviceFilterValue,
							},
						},
					},
				},
				{
					Dimensions: &keyValues{
						Key: "SERVICE",
						Values: []string{
							string(constant.AwsCostExplorer),
							string(constant.AwsTax),
						},
					},
				},
			},
		},
		// the order of group by will effect the result's key order
		Groups: []groupBy{
			{
				Key:  "SERVICE",
				Type: "DIMENSION",
			},
			{
				Key:  "RESOURCE_ID",
				Type: "DIMENSION",
			},
		},
	}

	m, err := json.Marshal(costWithResourceReq)
	if err != nil {
		return nil, err
	}

	res, err := a.sc.GetCostWithResourceWithContext(
		ctx,
		spider.AnycallReq{
			ConnectionName: param.ConnectionName,
			ReqInfo: spider.ReqInfo{
				FID: "getCostWithResource",
				IKeyValueList: []spider.KeyValue{
					{
						Key:   "requestBody",
						Value: string(m),
					},
				},
			},
		},
	)

	if err != nil {
		if errors.Is(err, spider.ErrSpiderCostResultEmpty) {
			log.Error().Msgf("error from spider: ", err)
			return nil, ErrCostResultEmpty
		}
		return nil, err
	}

	if res == nil || res.ResultsByTime == nil || len(res.ResultsByTime) == 0 {
		log.Error().Msgf("cost result is empty: ")
		return nil, ErrCostResultEmpty
	}

	var costInfos = make([]EstimateForecastCostInfo, 0)
	for _, result := range res.ResultsByTime {
		if result.Groups == nil {
			log.Error().Msgf("groups is nil; it must not be nil")
			return nil, ErrCostResultFormatInvalid
		}

		if result.TimePeriod == nil || result.TimePeriod.Start == nil || result.TimePeriod.End == nil {
			log.Error().Msgf("time period is nil; it must not be nil")
			return nil, ErrCostResultFormatInvalid
		}

		for _, group := range result.Groups {

			if group == nil {
				log.Error().Msgf("sinble group is nil; it must not be nil")
				continue
			}

			category := utils.NilSafeStr(group.Keys[0])
			awsService := constant.AwsService(category)
			resourceType, ok := serviceToResourceType[awsService]
			if !ok {
				log.Error().Msgf("service : %s does not exist; category: %s", awsService, category)
				continue
			}
			mt, ok := group.Metrics[metrics]
			if !ok {
				log.Error().Msgf("matric value does not exist:", metrics)
				continue
			}
			cost, err := strconv.ParseFloat(utils.NilSafeStr(mt.Amount), 64)
			if err != nil {
				log.Error().Msgf("cost parsing error:", mt.Amount)
				continue
			}
			unit := utils.NilSafeStr(mt.Unit)
			formattedResourceId := utils.NilSafeStr(group.Keys[1])
			actualResourceId := formattedResourceId

			if strings.Contains(formattedResourceId, "/") {
				splitedResource := strings.Split(actualResourceId, "/")
				actualResourceId = splitedResource[len(splitedResource)-1]
			}

			startDate, err := time.Parse(time.RFC3339, utils.NilSafeStr(result.TimePeriod.Start))
			if err != nil {
				log.Error().Msgf("start date parsing error:", result.TimePeriod.Start)
				continue
			}

			endDate, err := time.Parse(time.RFC3339, utils.NilSafeStr(result.TimePeriod.End))
			if err != nil {
				log.Error().Msgf("end date parsing error to ")
				continue
			}

			costInfo := EstimateForecastCostInfo{
				Provider:            param.Provider,
				ConnectionName:      param.ConnectionName,
				ResourceType:        resourceType,
				Category:            category,
				Cost:                cost,
				Unit:                unit,
				ActualResourceId:    actualResourceId,
				FormattedResourceId: formattedResourceId,
				Granularity:         granularity,
				StartDate:           startDate,
				EndDate:             endDate,
			}

			costInfos = append(costInfos, costInfo)
		}
	}

	return costInfos, nil
}

const (
	nsKey                      = "sys.namespace"
	provider                   = "aws"
	costExplorerConnectionName = "aws-us-east-1"
	defaultNsId                = "ns01"
	defaultMciId               = "mmci01"
)

func (a *AwsCostExplorerBaristaCostCollector) UpdateEstimateForecastCost(ctx context.Context, param UpdateEstimateForecastCostParam) (EstimateForecastCostInfos, error) {

	if param.NsId == "" {
		param.NsId = defaultNsId
	}

	if param.MciId == "" {
		param.MciId = defaultMciId
	}

	res := EstimateForecastCostInfos{}

	mci, err := a.tc.GetMciWithContext(ctx, param.NsId, param.MciId)

	if err != nil {
		log.Error().Msgf("error while get mci from tumblebug; ", err)
		return res, err
	}

	if len(mci.Vm) == 0 {
		return nil, errors.New("there is no vm in mci")
	}

	mciLabels := mci.Label
	_ = mciLabels[nsKey]

	arg := UpdateEstimateForecastCostRawParam{
		Provider:       provider,
		ConnectionName: costExplorerConnectionName,
		StartDate:      param.StartDate,
		EndDate:        param.EndDate,
		CostResources:  make([]CostResourceParam, 0),
	}

	vmIds := make([]string, 0)

	for _, mci := range mci.Vm {
		pn := mci.ConnectionConfig.ProviderName

		if pn == "" || !strings.EqualFold(strings.ToLower(pn), "aws") {
			log.Warn().Msgf("CSP: %s, does not support yet", pn)
			continue
		}

		vmId := mci.CspResourceId
		_ = mci.ConnectionName
		_ = mci.Label

		vmIds = append(vmIds, vmId)
	}

	if len(vmIds) != 0 {
		arg.CostResources = append(arg.CostResources, CostResourceParam{
			ResourceType: constant.VM,
			ResourceIds:  vmIds,
		})
	} else {
		return res, errors.New("no vm resource create on aws")
	}

	infos, err := a.GetCostInfos(ctx, arg)

	if err != nil {
		log.Error().Msgf("error while get cost info from spider;", err)
		return res, fmt.Errorf("error from get cost infos +%w", err)
	}

	for i := range infos {
		info := infos[i]
		info.NsId = param.NsId
		info.MciId = param.MciId
	}

	return infos, nil
}
