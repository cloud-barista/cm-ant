package spider

type ProductfamilyRes struct {
	Productfamily []string `json:"productfamily"`
}
type CloudPriceDataRes struct {
	Meta           MetaRes         `json:"meta"`
	CloudPriceList []CloudPriceRes `json:"cloudPriceList"`
	// Support v0.11.5 API response structure (PriceInfoResponse structure)
	CloudName  string     `json:"CloudName,omitempty"`
	RegionName string     `json:"RegionName,omitempty"`
	PriceList  []PriceRes `json:"PriceList,omitempty"`
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
	ProductInfo ProductInfoRes `json:"ProductInfo"`
	PriceInfo   PriceInfoRes   `json:"PriceInfo"`
	ZoneName    string         `json:"ZoneName,omitempty"`
}

type ProductInfoRes struct {
	ProductId  string `json:"productId"`
	RegionName string `json:"regionName"`
	ZoneName   string `json:"zoneName"`

	//--------- Compute Instance (v0.10.0 structure)
	InstanceType    string `json:"instanceType,omitempty"`
	Vcpu            string `json:"vcpu,omitempty"`
	Memory          string `json:"memory,omitempty"`
	Storage         string `json:"storage,omitempty"` // Root-Disk
	Gpu             string `json:"gpu,omitempty"`
	GpuMemory       string `json:"gpuMemory,omitempty"`
	OperatingSystem string `json:"operatingSystem,omitempty"`
	PreInstalledSw  string `json:"preInstalledSw,omitempty"`
	//--------- Compute Instance

	//--------- v0.11.5 new structure (exact field names)
	VMSpecInfo     *VMSpecInfoRes `json:"VMSpecInfo,omitempty"`
	Description    string         `json:"Description,omitempty"`
	CSPProductInfo interface{}    `json:"CSPProductInfo,omitempty"`
	//--------- v0.11.5 new structure

	//--------- Storage  // Data-Disk(AWS:EBS)
	VolumeType          string `json:"volumeType,omitempty"`
	StorageMedia        string `json:"storageMedia,omitempty"`
	MaxVolumeSize       string `json:"maxVolumeSize,omitempty"`
	MaxIOPSVolume       string `json:"maxIopsvolume,omitempty"`
	MaxThroughputVolume string `json:"maxThroughputvolume,omitempty"`
	//--------- Storage  // Data-Disk(AWS:EBS)
}

type VMSpecInfoRes struct {
	Region     string  `json:"Region"`
	Name       string  `json:"Name"`
	VCpu       VCpuRes `json:"VCpu"`
	MemSizeMiB string  `json:"MemSizeMiB"`
	DiskSizeGB string  `json:"DiskSizeGB"`
}

type VCpuRes struct {
	Count    string `json:"Count"`
	ClockGHz string `json:"ClockGHz"`
}

type PriceInfoRes struct {
	// cb-spider v0.11.5+ BREAKING: PricingPolicies []PricingPolicies -> OnDemand (단일 객체).
	// 신규 서버는 OnDemand 만 송신하므로 1순위 파싱 대상.
	// json tag도 v0.11.5+ 캡로 통일 (Go decoder는 case-insensitive이므로 기존 cspPriceInfo 도 호환).
	OnDemand        OnDemandRes          `json:"OnDemand"`
	CSPPriceInfo    interface{}          `json:"CSPPriceInfo"`
	// 옛 서버(<v0.11.5) 호환을 위한 leftover. v0.12.x 서버에선 비어 있다.
	PricingPolicies []PricingPoliciesRes `json:"pricingPolicies,omitempty"`
}

// OnDemandRes — cb-spider v0.11.5+ PriceInfo.OnDemand 정합.
// 모든 CSP 공통 단일 가격(시간당)이 직접 들어오므로 CSPPriceInfo 파싱 실패 시 fallback에 사용.
type OnDemandRes struct {
	PricingId   string `json:"PricingId"`
	Unit        string `json:"Unit"`
	Currency    string `json:"Currency"`
	Price       string `json:"Price"`
	Description string `json:"Description,omitempty"`
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
