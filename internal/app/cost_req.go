package app

import "github.com/cloud-barista/cm-ant/internal/core/common/constant"

type GetPriceInfoReq struct {
	ProviderName   string `json:"providerName" validate:"require"`
	ConnectionName string `json:"connectionName" validate:"require"`
	RegionName     string `json:"regionName" validate:"require"`
	InstanceType   string `json:"instanceType" validate:"require"`

	ZoneName string `json:"zoneName,omitempty"`
	VCpu     string `json:"vCpu,omitempty"`
	Memory   string `json:"memory,omitempty"`
	Storage  string `json:"storage,omitempty"`
	OsType   string `json:"osType,omitempty"`
}

type UpdateCostInfo struct {
	MigrationId       string            `json:"migrationId"`
	ConnectionName    string            `json:"connectionName"`
	StartDate         string            `json:"startDate" validate:"require"`
	EndDate           string            `json:"endDate" validate:"require"`
	CostResources     []CostResource    `json:"costResources" validate:"require"`
	AwsAdditionalInfo AwsAdditionalInfo `json:"awsAdditionalInfo"`
}

type CostResource struct {
	ResourceType constant.ResourceType `json:"resourceType"`
	ResourceIds  []string              `json:"resourceIds"`
}

type AwsAdditionalInfo struct {
	OwnerId string   `json:"ownerId"`
	Regions []string `json:"regions"`
}
