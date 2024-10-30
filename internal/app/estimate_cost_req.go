package app

import "github.com/cloud-barista/cm-ant/internal/core/common/constant"

type UpdateAndGetEstimateCostReq struct {
	Specs []struct {
		ProviderName string `json:"providerName" validate:"required"`
		RegionName   string `json:"regionName" validate:"required"`
		InstanceType string `json:"instanceType" validate:"required"`
		Image        string `json:"image"`
	} `json:"specs"`

	SpecsWithFormat []struct {
		CommonSpec  string `json:"commonSpec" validate:"required"`
		CommonImage string `json:"commonImage"`
	} `json:"specsWithFormat"`
}

type UpdateEstimateForecastCostReq struct {
	NsId  string `json:"nsId"`
	MciId string `json:"mciId"`
}

type GetEstimateCostInfosReq struct {
	ProviderName string `query:"providerName" validate:"required"`
	RegionName   string `query:"regionName" validate:"required"`
	InstanceType string `query:"instanceType"`
	VCpu         string `query:"vCpu"`
	Memory       string `query:"memory"`
	OsType       string `query:"osType"`
	Page         int    `query:"page"`
	Size         int    `query:"size"`
}

type GetEstimateForecastCostReq struct {
	Page                int                          `query:"page"`
	Size                int                          `query:"size"`
	StartDate           string                       `query:"startDate" validate:"required"`
	EndDate             string                       `query:"endDate" validate:"required"`
	NsIds               []string                     `query:"nsIds"`
	MciIds              []string                     `query:"mciIds"`
	Providers           []string                     `query:"provider"`
	ResourceTypes       []constant.ResourceType      `query:"resourceTypes"`
	ResourceIds         []string                     `query:"resourceIds"`
	CostAggregationType constant.CostAggregationType `query:"costAggregationType" validate:"required"`
	DateOrder           constant.OrderType           `query:"dateOrder"`
	ResourceTypeOrder   constant.OrderType           `query:"resourceTypeOrder"`
}

// -------------------------------------------------------------------------------------------------------------------

type UpdateCostInfoReq struct {
	CostResources     []CostResourceReq    `json:"costResources" validate:"required"`
	AwsAdditionalInfo AwsAdditionalInfoReq `json:"awsAdditionalInfo"`
}

type CostResourceReq struct {
	ResourceType constant.ResourceType `json:"resourceType"`
	ResourceIds  []string              `json:"resourceIds"`
}

type AwsAdditionalInfoReq struct {
	OwnerId string   `json:"ownerId"`
	Regions []string `json:"regions"`
}
