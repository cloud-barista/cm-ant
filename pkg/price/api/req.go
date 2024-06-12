package api

// PriceInfoReq is
type PriceInfoReq struct {
	ProviderName string `json:"providerName,omitempty"`

	ConnectionName string `json:"connectionName,omitempty"`
	RegionName     string `json:"regionName,omitempty"`
	VCpu           string `json:"vCpu,omitempty"`
	MemoryGiB      string `json:"memoryGiB,omitempty"`
	CspSpecName    string `json:"cspSpecName,omitempty"`
	OsType         string `json:"osType,omitempty"`

	RootDiskType string `json:"rootDiskType,omitempty"`
	RootDiskSize string `json:"rootDiskSize,omitempty"`
	StorageGiB   string `json:"storageGiB,omitempty"`

	AcceleratorCount    string `json:"acceleratorCount,omitempty"`
	AcceleratorMemoryGB string `json:"acceleratorMemoryGB,omitempty"`
	AcceleratorModel    string `json:"acceleratorModel,omitempty"`
	AcceleratorType     string `json:"acceleratorType,omitempty"`
}
