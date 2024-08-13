package spider

type ProductfamilyRes struct {
	Productfamily []string `json:"productfamily"`
}
type CloudPriceDataRes struct {
	Meta           MetaRes         `json:"meta"`
	CloudPriceList []CloudPriceRes `json:"cloudPriceList"`
}

type MetaRes struct {
	Version     string `json:"version"`
	Description string `json:"description"`
}

type CloudPriceRes struct {
	CloudName string     `json:"cloudName"`
	PriceList []PriceRes `json:"priceList"`
}

type PriceRes struct {
	ProductInfo ProductInfoRes `json:"productInfo"`
	PriceInfo   PriceInfoRes   `json:"priceInfo"`
}

type ProductInfoRes struct {
	ProductId  string `json:"productId"`
	RegionName string `json:"regionName"`
	ZoneName   string `json:"zoneName"`

	//--------- Compute Instance
	InstanceType    string `json:"instanceType,omitempty"`
	Vcpu            string `json:"vcpu,omitempty"`
	Memory          string `json:"memory,omitempty"`
	Storage         string `json:"storage,omitempty"` // Root-Disk
	Gpu             string `json:"gpu,omitempty"`
	GpuMemory       string `json:"gpuMemory,omitempty"`
	OperatingSystem string `json:"operatingSystem,omitempty"`
	PreInstalledSw  string `json:"preInstalledSw,omitempty"`
	//--------- Compute Instance

	//--------- Storage  // Data-Disk(AWS:EBS)
	VolumeType          string `json:"volumeType,omitempty"`
	StorageMedia        string `json:"storageMedia,omitempty"`
	MaxVolumeSize       string `json:"maxVolumeSize,omitempty"`
	MaxIOPSVolume       string `json:"maxIopsvolume,omitempty"`
	MaxThroughputVolume string `json:"maxThroughputvolume,omitempty"`
	//--------- Storage  // Data-Disk(AWS:EBS)

	Description    string      `json:"description"`
	CSPProductInfo interface{} `json:"cspProductInfo"`
}

type PriceInfoRes struct {
	PricingPolicies []PricingPoliciesRes `json:"pricingPolicies"`
	CSPPriceInfo    interface{}          `json:"cspPriceInfo"`
}

type PricingPoliciesRes struct {
	PricingId         string                `json:"pricingId"`
	PricingPolicy     string                `json:"pricingPolicy"`
	Unit              string                `json:"unit"`
	Currency          string                `json:"currency"`
	Price             string                `json:"price"`
	Description       string                `json:"description"`
	PricingPolicyInfo *PricingPolicyInfoRes `json:"pricingPolicyInfo,omitempty"`
}

type PricingPolicyInfoRes struct {
	LeaseContractLength string `json:"LeaseContractLength"`
	OfferingClass       string `json:"OfferingClass"`
	PurchaseOption      string `json:"PurchaseOption"`
}

type AnycallRes struct {
	FID           string     `json:"FID"`
	IKeyValueList []KeyValue `json:"IKeyValueList"`
	OKeyValueList []KeyValue `json:"OKeyValueList"`
}

type CostWithResourcesRes struct {
	DimensionValueAttributes []*DimensionValuesWithAttributes `json:"DimensionValueAttributes"`
	GroupDefinitions         []*GroupDefinition               `json:"GroupDefinitions"`
	NextPageToken            *string                          `json:"NextPageToken"`
	ResultsByTime            []*ResultByTime                  `json:"ResultsByTime"`
}

type DimensionValuesWithAttributes struct {
	Attributes map[string]*string `json:"Attributes"`
	Value      *string            `json:"Value"`
}

type GroupDefinition struct {
	Key  *string `json:"Key"`
	Type *string `json:"Type"`
}

type ResultByTime struct {
	Estimated  *bool                   `json:"Estimated"`
	Groups     []*Group                `json:"Groups"`
	TimePeriod *DateInterval           `json:"TimePeriod"`
	Total      map[string]*MetricValue `json:"Total"`
}

type Group struct {
	Keys    []*string               `json:"Keys"`
	Metrics map[string]*MetricValue `json:"Metrics"`
}

type DateInterval struct {
	End   *string `json:"End"`
	Start *string `json:"Start"`
}

type MetricValue struct {
	Amount *string `json:"Amount"`
	Unit   *string `json:"Unit"`
}
