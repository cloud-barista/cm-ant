package cost

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
)

type EstimateForecastCostParam struct {
	RecommendSpecs []RecommendSpecParam `json:"recommendSpecs"`

	TimeStandard time.Time            `json:"timeStandard"`
	PricePolicy  constant.PricePolicy `json:"pricePolicy"`
}

type RecommendSpecParam struct {
	ProviderName string `json:"providerName"`
	RegionName   string `json:"regionName"`
	InstanceType string `json:"instanceType"`
	Image        string `json:"image"`
}

func (r RecommendSpecParam) Hash() string {
	h := sha256.New()

	h.Write([]byte(r.ProviderName))
	h.Write([]byte(r.RegionName))
	h.Write([]byte(r.InstanceType))
	h.Write([]byte(r.Image))

	hashBytes := h.Sum(nil)
	return hex.EncodeToString(hashBytes)
}

type EstimateForecastCostResult struct {
	TotalMinMonthlyPrice           float64                         `json:"totalMinMonthlyPrice"`
	TotalMaxMonthlyPrice           float64                         `json:"totalMaxMonthlyPrice"`
	EsimateForecastCostSpecResults []EsimateForecastCostSpecResult `json:"esimateForecastCostSpecResults"`
}

type EsimateForecastCostSpecResult struct {
	ProviderName                          string                                 `json:"providerName"`
	RegionName                            string                                 `json:"regionName"`
	InstanceType                          string                                 `json:"instanceType"`
	ImageName                             string                                 `json:"imageName"`
	SpecMinMonthlyPrice                   float64                                `json:"totalMinMonthlyPrice"`
	SpecMaxMonthlyPrice                   float64                                `json:"totalMaxMonthlyPrice"`
	EstimateForecastCostSpecDetailResults []EstimateForecastCostSpecDetailResult `json:"estimateForecastCostSpecDetailResults"`
}

type EstimateForecastCostSpecDetailResult struct {
	ID                     uint                   `json:"id"`
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
	CalculatedMonthlyPrice float64                `json:"calculatedMonthlyPrice,omitempty"`
	PriceDescription       string                 `json:"priceDescription,omitempty"`
	LastUpdatedAt          time.Time              `json:"lastUpdatedAt,omitempty"`
}

type UpdatePriceInfosParam struct {
	MigrationId  string
	ProviderName string
	RegionName   string
	InstanceType string

	TimeStandard time.Time
	PricePolicy  constant.PricePolicy
}

type GetPriceInfosParam struct {
	ProviderName string
	RegionName   string
	InstanceType string

	VCpu   string
	Memory string
	OsType string

	TimeStandard time.Time
	PricePolicy  constant.PricePolicy
}

type AllPriceInfoResult struct {
	PriceInfoList []PriceInfoResult `json:"priceInfoList,omitempty"`
	ResultCount   int64             `json:"resultCount"`
}

type PriceInfoResult struct {
	ID           uint   `json:"id"`
	ProviderName string `json:"providerName"`
	RegionName   string `json:"regionName"`
	InstanceType string `json:"instanceType"`

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
	CalculatedMonthlyPrice float64                `json:"calculatedMonthlyPrice,omitempty"`
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
