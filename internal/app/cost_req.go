package app

import "github.com/cloud-barista/cm-ant/internal/core/common/constant"

type GetPriceInfoReq struct {
	ProviderName   string `query:"providerName" validate:"required"`
	ConnectionName string `query:"connectionName" validate:"required"`
	RegionName     string `query:"regionName" validate:"required"`
	InstanceType   string `query:"instanceType" validate:"required"`

	ZoneName string `query:"zoneName,omitempty"`
	VCpu     string `query:"vCpu,omitempty"`
	Memory   string `query:"memory,omitempty"`
	Storage  string `query:"storage,omitempty"`
	OsType   string `query:"osType,omitempty"`
}

type UpdateCostInfoReq struct {
	MigrationId       string               `json:"migrationId"`
	ConnectionName    string               `json:"connectionName" validate:"required"`
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

type GetCostInfoReq struct {
	StartDate           string                       `query:"startDate" validate:"required"`
	EndDate             string                       `query:"endDate" validate:"required"`
	MigrationIds        []string                     `query:"migrationIds"`
	Providers           []string                     `query:"provider"`
	ResourceTypes       []constant.ResourceType      `query:"resourceTypes"`
	ResourceIds         []string                     `query:"resourceIds"`
	CostAggregationType constant.CostAggregationType `query:"costAggregationType" validate:"required"`
	DateOrder           constant.OrderType           `query:"dateOrder"`
	ResourceTypeOrder   constant.OrderType           `query:"resourceTypeOrder"`
}
