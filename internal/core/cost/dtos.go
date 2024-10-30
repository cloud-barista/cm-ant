package cost

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
)

type UpdateAndGetEstimateCostParam struct {
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

type EstimateCostResults struct {
	EsimateCostSpecResults []EsimateCostSpecResults `json:"esimateCostSpecResults,omitempty"`
}

type EsimateCostSpecResults struct {
	ProviderName                  string                         `json:"providerName,omitempty"`
	RegionName                    string                         `json:"regionName,omitempty"`
	InstanceType                  string                         `json:"instanceType,omitempty"`
	ImageName                     string                         `json:"imageName,omitempty"`
	SpecMinMonthlyPrice           float64                        `json:"totalMinMonthlyPrice,omitempty"`
	SpecMaxMonthlyPrice           float64                        `json:"totalMaxMonthlyPrice,omitempty"`
	EstimateCostSpecDetailResults []EstimateCostSpecDetailResult `json:"estimateForecastCostSpecDetailResults,omitempty"`
}

type EstimateCostSpecDetailResult struct {
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

type GetEstimateCostParam struct {
	ProviderName string
	RegionName   string
	InstanceType string

	VCpu   string
	Memory string
	OsType string

	TimeStandard time.Time
	PricePolicy  constant.PricePolicy
	Page         int
	Size         int
}

type EstimateCostInfoResults struct {
	EstimateCostInfoResult []EstimateCostInfoResult `json:"estimateCostInfoResult,omitempty"`
	ResultCount            int64                    `json:"resultCount"`
}

type EstimateCostInfoResult struct {
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

type UpdateEstimateForecastCostParam struct {
	NsId      string
	MciId     string
	StartDate time.Time
	EndDate   time.Time
}

type UpdateEstimateForecastCostInfoResult struct {
	FetchedDataCount  int64 `json:"fetchedDataCount"`
	UpdatedDataCount  int64 `json:"updatedDataCount"`
	InsertedDataCount int64 `json:"insertedDataCount"`
}

type GetEstimateForecastCostParam struct {
	Page                int
	Size                int
	StartDate           time.Time
	EndDate             time.Time
	NsIds               []string
	MciIds              []string
	Providers           []string
	ResourceTypes       []constant.ResourceType
	ResourceIds         []string
	CostAggregationType constant.CostAggregationType
	DateOrder           constant.OrderType
	ResourceTypeOrder   constant.OrderType
}

type GetEstimateForecastCostInfoResults struct {
	GetEstimateForecastCostInfoResults []GetEstimateForecastCostInfoResult `json:"getEstimateForecastCostInfoResults,omitempty"`
	ResultCount                        int64                               `json:"resultCount"`
}
type GetEstimateForecastCostInfoResult struct {
	Provider         string    `json:"provider"`
	ResourceType     string    `json:"resourceType"`
	Category         string    `json:"category"`
	ActualResourceId string    `json:"resourceId"`
	Unit             string    `json:"unit"`
	Date             time.Time `json:"date"`
	TotalCost        float64   `json:"totalCost"`
}

// -------------------------------------------------------------------

type UpdateEstimateForecastCostRawParam struct {
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
