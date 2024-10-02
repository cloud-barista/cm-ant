package tumblebug

type MciRes struct {
	ResourceType                  string            `json:"resourceType"`
	Id                            string            `json:"id"`
	Uid                           string            `json:"uid"`
	Name                          string            `json:"name"`
	Status                        string            `json:"status"`
	StatusCount                   StatusCountRes    `json:"statusCount"`
	TargetStatus                  string            `json:"targetStatus"`
	TargetAction                  string            `json:"targetAction"`
	InstallMonAgent               string            `json:"installMonAgent"`
	ConfigureCloudAdaptiveNetwork string            `json:"configureCloudAdaptiveNetwork"`
	Label                         map[string]string `json:"label"`
	SystemLabel                   string            `json:"systemLabel"`
	SystemMessage                 string            `json:"systemMessage"`
	Description                   string            `json:"description"`
	Vm                            []VmRes           `json:"vm"`
	NewVMList                     []string          `json:"newVmList"`
}

type StatusCountRes struct {
	CountTotal       int `json:"countTotal,omitempty"`
	CountCreating    int `json:"countCreating,omitempty"`
	CountRunning     int `json:"countRunning,omitempty"`
	CountFailed      int `json:"countFailed,omitempty"`
	CountSuspended   int `json:"countSuspended,omitempty"`
	CountRebooting   int `json:"countRebooting,omitempty"`
	CountTerminated  int `json:"countTerminated,omitempty"`
	CountSuspending  int `json:"countSuspending,omitempty"`
	CountResuming    int `json:"countResuming,omitempty"`
	CountTerminating int `json:"countTerminating,omitempty"`
	CountUndefined   int `json:"countUndefined,omitempty"`
}

type RegionRes struct {
	Region string `json:"Region,omitempty"`
	Zone   string `json:"Zone,omitempty"`
}
type ConnectionConfigRes struct {
	ConfigName         string `json:"configName"`
	ProviderName       string `json:"providerName"`
	DriverName         string `json:"driverName"`
	CredentialName     string `json:"credentialName"`
	CredentialHolder   string `json:"credentialHolder"`
	RegionZoneInfoName string `json:"regionZoneInfoName"`
	RegionZoneInfo     struct {
		AssignedRegion string `json:"assignedRegion"`
		AssignedZone   string `json:"assignedZone"`
	} `json:"regionZoneInfo"`
	RegionDetail struct {
		RegionID    string `json:"regionId"`
		RegionName  string `json:"regionName"`
		Description string `json:"description"`
		Location    struct {
			Display   string  `json:"display"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"location"`
		Zones []string `json:"zones"`
	} `json:"regionDetail"`
	RegionRepresentative bool `json:"regionRepresentative"`
	Verified             bool `json:"verified"`
}

type VmRes struct {
	ResourceType    string `json:"resourceType"`
	Id              string `json:"id"`
	Uid             string `json:"uid"`
	CspResourceName string `json:"cspResourceName"`
	CspResourceId   string `json:"cspResourceId"`
	Name            string `json:"name"`
	SubGroupId      string `json:"subGroupId"`
	Location        struct {
		Display   string  `json:"display"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"location"`
	Status             string              `json:"status"`
	TargetStatus       string              `json:"targetStatus"`
	TargetAction       string              `json:"targetAction"`
	MonAgentStatus     string              `json:"monAgentStatus"`
	NetworkAgentStatus string              `json:"networkAgentStatus"`
	SystemMessage      string              `json:"systemMessage"`
	CreatedTime        string              `json:"createdTime"`
	Label              map[string]string   `json:"label"`
	Description        string              `json:"description"`
	Region             RegionRes           `json:"region"`
	PublicIP           string              `json:"publicIP"`
	SSHPort            string              `json:"sshPort"`
	PublicDNS          string              `json:"publicDNS"`
	PrivateIP          string              `json:"privateIP"`
	PrivateDNS         string              `json:"privateDNS"`
	RootDiskType       string              `json:"rootDiskType"`
	RootDiskSize       string              `json:"rootDiskSize"`
	RootDeviceName     string              `json:"rootDeviceName"`
	ConnectionName     string              `json:"connectionName"`
	ConnectionConfig   ConnectionConfigRes `json:"connectionConfig"`
	SpecId             string              `json:"specId"`
	CspSpecName        string              `json:"cspSpecName"`
	ImageId            string              `json:"imageId"`
	CspImageName       string              `json:"cspImageName"`
	VNetID             string              `json:"vNetId"`
	CspVNetId          string              `json:"cspVNetId"`
	SubnetId           string              `json:"subnetId"`
	CspSubnetId        string              `json:"cspSubnetId"`
	NetworkInterface   string              `json:"networkInterface"`
	SecurityGroupIds   []string            `json:"securityGroupIds"`
	DataDiskIds        []string            `json:"dataDiskIds"`
	SSHKeyId           string              `json:"sshKeyId"`
	CspSSHKeyId        string              `json:"cspSshKeyId"`
	VMUserName         string              `json:"vmUserName"`
	AddtionalDetails   []KeyValueRes       `json:"addtionalDetails"`
}

type KeyValueRes struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type RecommendVmResList []RecommendVmRes

type RecommendVmRes struct {
	AcceleratorCount      int      `json:"acceleratorCount"`
	AcceleratorMemoryGB   int      `json:"acceleratorMemoryGB"`
	AcceleratorModel      string   `json:"acceleratorModel"`
	AcceleratorType       string   `json:"acceleratorType"`
	AssociatedObjectList  []string `json:"associatedObjectList"`
	Description           string   `json:"description"`
	EvaluationStatus      string   `json:"evaluationStatus"`
	InfraType             string   `json:"infraType"`
	IsAutoGenerated       bool     `json:"isAutoGenerated"`
	MaxTotalStorageTiB    int      `json:"maxTotalStorageTiB"`
	NetBwGbps             int      `json:"netBwGbps"`
	OsType                string   `json:"osType"`
	StorageGiB            int      `json:"storageGiB"`
	SystemLabel           string   `json:"systemLabel"`
	Uid                   string   `json:"uid"`
	Namespace             string   `json:"namespace"`
	Id                    string   `json:"id"`
	Name                  string   `json:"name"`
	ConnectionName        string   `json:"connectionName"`
	ProviderName          string   `json:"providerName"`
	RegionName            string   `json:"regionName"`
	CspSpecName           string   `json:"cspSpecName"`
	VCPU                  int      `json:"vCPU"`
	MemoryGiB             int      `json:"memoryGiB"`
	CostPerHour           float64  `json:"costPerHour"`
	OrderInFilteredResult int      `json:"orderInFilteredResult"`
	EvaluationScore01     float64  `json:"evaluationScore01"`
	EvaluationScore02     int      `json:"evaluationScore02"`
	EvaluationScore03     int      `json:"evaluationScore03"`
	EvaluationScore04     int      `json:"evaluationScore04"`
	EvaluationScore05     int      `json:"evaluationScore05"`
	EvaluationScore06     int      `json:"evaluationScore06"`
	EvaluationScore07     int      `json:"evaluationScore07"`
	EvaluationScore08     int      `json:"evaluationScore08"`
	EvaluationScore09     int      `json:"evaluationScore09"`
	EvaluationScore10     float64  `json:"evaluationScore10"`
	RootDiskType          string   `json:"rootDiskType"`
	RootDiskSize          string   `json:"rootDiskSize"`
}

type GetNsRes struct {
	ResourceType string `json:"resourceType"`
	Uid          string `json:"uid"`
	Id           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
}
