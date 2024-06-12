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
