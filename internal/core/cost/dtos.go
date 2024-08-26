package cost

import (
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
)

type GetPriceInfoParam struct {
	MigrationId  string
	ProviderName string
	RegionName   string
	InstanceType string
	ZoneName     string
	VCpu         string
	Memory       string
	Storage      string
	OsType       string
	Unit         constant.PriceUnit
	Currency     constant.PriceCurrency

	TimeStandard time.Time
	PricePolicy  constant.PricePolicy
}

type AllPriceInfoResult struct {
	PriceInfoList []PriceInfoResult `json:"priceInfoList,omitempty"`
	InfoSource    string            `json:"infoSource,omitempty"`
	ResultCount   int64             `json:"resultCount"`
}

type PriceInfoResult struct {
	ID           uint   `json:"id"`
	ProviderName string `json:"providerName"`
	RegionName   string `json:"regionName"`
	InstanceType string `json:"instanceType"`

	ZoneName               string                 `json:"zoneName,omitempty"`
	VCpu                   string                 `json:"vCpu,omitempty"`
	Memory                 string                 `json:"memory,omitempty"`
	Storage                string                 `json:"storage,omitempty"`
	OsType                 string                 `json:"osType,omitempty"`
	ProductDescription     string                 `json:"productDescription,omitempty"`
	OriginalPricePolicy    string                 `json:"originalPricePolicy,omitempty"`
	PricePolicy            constant.PricePolicy   `json:"pricePolicy,omitempty"`
	Unit                   constant.PriceUnit     `json:"unit,omitempty"`
	Currency               constant.PriceCurrency `json:"currency,omitempty"`
	Price                  string                 `json:"price,omitempty"`
	CalculatedMonthlyPrice string                 `json:"calculatedMonthlyPrice,omitempty"`
	PriceDescription       string                 `json:"priceDescription,omitempty"`
	LastUpdatedAt          time.Time              `json:"lastUpdatedAt,omitempty"`
}

type UpdateCostInfoParam struct {
	MigrationId       string
	Provider          string // currently only aws
	ConnectionName    string
	StartDate         time.Time
	EndDate           time.Time
	CostResources     []CostResourceParam
	AwsAdditionalInfo AwsAdditionalInfoParam
}

type CostResourceParam struct {
	ResourceType constant.ResourceType
	ResourceIds  []string
}

type AwsAdditionalInfoParam struct {
	OwnerId string   `json:"ownerId"`
	Regions []string `json:"regions"`
}

type UpdateCostInfoResult struct {
	FetchedDataCount  int64 `json:"fetchedDataCount"`
	UpdatedDataCount  int64 `json:"updatedDataCount"`
	InsertedDataCount int64 `insertedDataCount`
}

type GetCostInfoParam struct {
	StartDate           time.Time
	EndDate             time.Time
	MigrationIds        []string
	Providers           []string
	ResourceTypes       []constant.ResourceType
	ResourceIds         []string
	CostAggregationType constant.CostAggregationType
	DateOrder           constant.OrderType
	ResourceTypeOrder   constant.OrderType
}

type GetCostInfoResult struct {
	Provider         string    `json:"provider"`
	ResourceType     string    `json:"resourceType"`
	Category         string    `json:"category"`
	ActualResourceId string    `json:"resourceId"`
	Unit             string    `json:"unit"`
	Date             time.Time `json:"date"`
	TotalCost        float64   `json:"totalCost"`
}
