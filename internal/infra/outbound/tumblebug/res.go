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
	SystemMessage                 []string          `json:"systemMessage"`
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

// CB-Tumblebug v0.11.8+ 새로운 응답 구조체
type SpecInfoList []SpecInfo

type SpecInfo struct {
	Id                    string   `json:"id"`
	Uid                   string   `json:"uid,omitempty"`
	CspSpecName           string   `json:"cspSpecName,omitempty"`
	Name                  string   `json:"name"`
	Namespace             string   `json:"namespace,omitempty"`
	ConnectionName        string   `json:"connectionName,omitempty"`
	ProviderName          string   `json:"providerName,omitempty"`
	RegionName            string   `json:"regionName,omitempty"`
	RegionLatitude        float64  `json:"regionLatitude"`
	RegionLongitude       float64  `json:"regionLongitude"`
	InfraType             string   `json:"infraType,omitempty"`
	Architecture          string   `json:"architecture,omitempty"`
	OsType                string   `json:"osType,omitempty"`
	VCPU                  uint16   `json:"vCPU,omitempty"`
	MemoryGiB             float32  `json:"memoryGiB,omitempty"`
	DiskSizeGB            float32  `json:"diskSizeGB,omitempty"`
	MaxTotalStorageTiB    uint16   `json:"maxTotalStorageTiB,omitempty"`
	NetBwGbps             uint16   `json:"netBwGbps,omitempty"`
	AcceleratorModel      string   `json:"acceleratorModel,omitempty"`
	AcceleratorCount      uint8    `json:"acceleratorCount,omitempty"`
	AcceleratorMemoryGB   float32  `json:"acceleratorMemoryGB,omitempty"`
	AcceleratorType       string   `json:"acceleratorType,omitempty"`
	CostPerHour           float32  `json:"costPerHour,omitempty"`
	Description           string   `json:"description,omitempty"`
	OrderInFilteredResult uint16   `json:"orderInFilteredResult,omitempty"`
	EvaluationStatus      string   `json:"evaluationStatus,omitempty"`
	EvaluationScore01     float32  `json:"evaluationScore01"`
	EvaluationScore02     float32  `json:"evaluationScore02"`
	EvaluationScore03     float32  `json:"evaluationScore03"`
	EvaluationScore04     float32  `json:"evaluationScore04"`
	EvaluationScore05     float32  `json:"evaluationScore05"`
	EvaluationScore06     float32  `json:"evaluationScore06"`
	EvaluationScore07     float32  `json:"evaluationScore07"`
	EvaluationScore08     float32  `json:"evaluationScore08"`
	EvaluationScore09     float32  `json:"evaluationScore09"`
	EvaluationScore10     float32  `json:"evaluationScore10"`
	RootDiskType          string   `json:"rootDiskType"`
	RootDiskSize          string   `json:"rootDiskSize"`
	AssociatedObjectList  []string `json:"associatedObjectList,omitempty"`
	IsAutoGenerated       bool     `json:"isAutoGenerated,omitempty"`
	SystemLabel           string   `json:"systemLabel,omitempty"`
}

// 기존 호환성을 위한 별칭 (deprecated)
type RecommendVmResList = SpecInfoList
type RecommendVmRes = SpecInfo

// CB-Tumblebug 이미지 정보 구조체
type ImageInfo struct {
	Id                   string   `json:"id"`
	Uid                  string   `json:"uid,omitempty"`
	Name                 string   `json:"name"`
	ConnectionName       string   `json:"connectionName,omitempty"`
	CspImageId           string   `json:"cspImageId,omitempty"`
	CspImageName         string   `json:"cspImageName,omitempty"`
	Description          string   `json:"description,omitempty"`
	GuestOS              string   `json:"guestOS,omitempty"`
	Status               string   `json:"status,omitempty"`
	KeyValueList         []string `json:"keyValueList,omitempty"`
	AssociatedObjectList []string `json:"associatedObjectList,omitempty"`
	IsAutoGenerated      bool     `json:"isAutoGenerated,omitempty"`
	SystemLabel          string   `json:"systemLabel,omitempty"`
}

type GetNsRes struct {
	ResourceType string `json:"resourceType"`
	Uid          string `json:"uid"`
	Id           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
}

// SshCmdResultForAPI is struct for SshCmd Result with string error for API response
type SshCmdResultForAPI struct {
	MciId   string         `json:"mciId"`
	VmId    string         `json:"vmId"`
	VmIp    string         `json:"vmIp"`
	Command map[int]string `json:"command"`
	Stdout  map[int]string `json:"stdout"`
	Stderr  map[int]string `json:"stderr"`
	Error   string         `json:"error"` // String representation of error for JSON serialization
}

// MciSshCmdResultForAPI is struct for Set of SshCmd Results in terms of MCI for API response
type MciSshCmdResultForAPI struct {
	Results []SshCmdResultForAPI `json:"results"`
}
