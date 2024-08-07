package app

type PriceInfoReq struct {
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
