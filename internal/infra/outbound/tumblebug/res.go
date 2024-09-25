package tumblebug

import (
	"time"
)

type MciRes struct {
	ID                            string            `json:"id,omitempty"`
	Name                          string            `json:"name,omitempty"`
	Status                        string            `json:"status,omitempty"`
	StatusCount                   StatusCountRes    `json:"statusCount,omitempty"`
	TargetStatus                  string            `json:"targetStatus,omitempty"`
	TargetAction                  string            `json:"targetAction,omitempty"`
	InstallMonAgent               string            `json:"installMonAgent,omitempty"`
	ConfigureCloudAdaptiveNetwork string            `json:"configureCloudAdaptiveNetwork,omitempty"`
	Label                         map[string]string `json:"label,omitempty"`
	SystemLabel                   string            `json:"systemLabel,omitempty"`
	SystemMessage                 string            `json:"systemMessage,omitempty"`
	Description                   string            `json:"description,omitempty"`
	VMs                           []VmRes           `json:"vm,omitempty"`
	NewVMList                     []string          `json:"newVmList,omitempty"`
	PlacementAlgo                 string            `json:"placementAlgo"`
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
type LocationRes struct {
	Latitude     float64 `json:"latitude,omitempty"`
	Longitude    float64 `json:"longitude,omitempty"`
	BriefAddr    string  `json:"briefAddr,omitempty"`
	CloudType    string  `json:"cloudType,omitempty"`
	NativeRegion string  `json:"nativeRegion,omitempty"`
}
type RegionRes struct {
	Region string `json:"Region,omitempty"`
	Zone   string `json:"Zone,omitempty"`
}
type ConnectionConfigRes struct {
	ConfigName     string      `json:"ConfigName,omitempty"`
	ProviderName   string      `json:"ProviderName,omitempty"`
	DriverName     string      `json:"DriverName,omitempty"`
	CredentialName string      `json:"CredentialName,omitempty"`
	RegionName     string      `json:"RegionName,omitempty"`
	Location       LocationRes `json:"Location,omitempty"`
}

type IIDRes struct {
	NameID   string `json:"NameId,omitempty"`
	SystemID string `json:"SystemId,omitempty"`
}

type CspViewDetailRes struct {
	Name               string        `json:"Name,omitempty"`
	ImageName          string        `json:"ImageName,omitempty"`
	VPCName            string        `json:"VPCName,omitempty"`
	SubnetName         string        `json:"SubnetName,omitempty"`
	SecurityGroupNames []string      `json:"SecurityGroupNames,omitempty"`
	KeyPairName        string        `json:"KeyPairName,omitempty"`
	CSPid              string        `json:"CSPid,omitempty"`
	DataDiskNames      []string      `json:"DataDiskNames,omitempty"`
	VMSpecName         string        `json:"VMSpecName,omitempty"`
	VMUserID           string        `json:"VMUserId,omitempty"`
	VMUserPasswd       string        `json:"VMUserPasswd,omitempty"`
	RootDiskType       string        `json:"RootDiskType,omitempty"`
	RootDiskSize       string        `json:"RootDiskSize,omitempty"`
	ImageType          string        `json:"ImageType,omitempty"`
	IID                IIDRes        `json:"IId,omitempty"`
	ImageIID           IIDRes        `json:"ImageIId,omitempty"`
	VpcIID             IIDRes        `json:"VpcIID,omitempty"`
	SubnetIID          IIDRes        `json:"SubnetIID,omitempty"`
	SecurityGroupIIds  []IIDRes      `json:"SecurityGroupIIds,omitempty"`
	KeyPairIID         IIDRes        `json:"KeyPairIId,omitempty"`
	DataDiskIIDs       []IIDRes      `json:"DataDiskIIDs,omitempty"`
	StartTime          time.Time     `json:"StartTime,omitempty"`
	Region             RegionRes     `json:"Region,omitempty"`
	NetworkInterface   string        `json:"NetworkInterface,omitempty"`
	PublicIP           string        `json:"PublicIP,omitempty"`
	PublicDNS          string        `json:"PublicDNS,omitempty"`
	PrivateIP          string        `json:"PrivateIP,omitempty"`
	PrivateDNS         string        `json:"PrivateDNS,omitempty"`
	RootDeviceName     string        `json:"RootDeviceName,omitempty"`
	SSHAccessPoint     string        `json:"SSHAccessPoint,omitempty"`
	KeyValueList       []KeyValueRes `json:"KeyValueList,omitempty"`
}
type VmRes struct {
	ID                 string              `json:"id,omitempty"`
	Name               string              `json:"name,omitempty"`
	IDByCSP            string              `json:"idByCSP,omitempty"`
	SubGroupID         string              `json:"subGroupId,omitempty"`
	Location           LocationRes         `json:"location,omitempty"`
	Status             string              `json:"status,omitempty"`
	TargetStatus       string              `json:"targetStatus,omitempty"`
	TargetAction       string              `json:"targetAction,omitempty"`
	MonAgentStatus     string              `json:"monAgentStatus,omitempty"`
	NetworkAgentStatus string              `json:"networkAgentStatus,omitempty"`
	SystemMessage      string              `json:"systemMessage,omitempty"`
	CreatedTime        string              `json:"createdTime,omitempty"`
	Label              map[string]string   `json:"label,omitempty"`
	Description        string              `json:"description,omitempty"`
	Region             RegionRes           `json:"region,omitempty"`
	PublicIP           string              `json:"publicIP,omitempty"`
	SSHPort            string              `json:"sshPort,omitempty"`
	PublicDNS          string              `json:"publicDNS,omitempty"`
	PrivateIP          string              `json:"privateIP,omitempty"`
	PrivateDNS         string              `json:"privateDNS,omitempty"`
	RootDiskType       string              `json:"rootDiskType,omitempty"`
	RootDiskSize       string              `json:"rootDiskSize,omitempty"`
	RootDeviceName     string              `json:"rootDeviceName,omitempty"`
	ConnectionName     string              `json:"connectionName,omitempty"`
	ConnectionConfig   ConnectionConfigRes `json:"connectionConfig,omitempty"`
	SpecID             string              `json:"specId,omitempty"`
	ImageID            string              `json:"imageId,omitempty"`
	VNetID             string              `json:"vNetId,omitempty"`
	SubnetID           string              `json:"subnetId,omitempty"`
	SecurityGroupIds   []string            `json:"securityGroupIds,omitempty"`
	DataDiskIds        []string            `json:"dataDiskIds,omitempty"`
	SSHKeyID           string              `json:"sshKeyId,omitempty"`
	VMUserAccount      string              `json:"vmUserAccount,omitempty"`
	CspViewVMDetail    CspViewDetailRes    `json:"cspViewVmDetail,omitempty"`
}

type SecurityGroupRes struct {
	AssociatedObjectList []string          `json:"associatedObjectList"`
	ConnectionName       string            `json:"connectionName"`
	CspSecurityGroupId   string            `json:"cspSecurityGroupId"`
	CspSecurityGroupName string            `json:"cspSecurityGroupName"`
	Description          string            `json:"description"`
	FirewallRules        []FirewallRuleRes `json:"firewallRules"`
	Id                   string            `json:"id"`
	IsAutoGenerated      bool              `json:"isAutoGenerated"`
	KeyValueList         []KeyValueRes     `json:"keyValueList"`
	Name                 string            `json:"name"`
	SystemLabel          string            `json:"systemLabel"`
	VNetId               string            `json:"vNetId"`
}

type FirewallRuleRes struct {
	FromPort   string `json:"FromPort"`
	ToPort     string `json:"ToPort"`
	IPProtocol string `json:"IPProtocol"`
	Direction  string `json:"Direction"`
	CIDR       string `json:"CIDR"`
}

type KeyValueRes struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SecureShellRes struct {
	AssociatedObjectList []string      `json:"associatedObjectList"`
	ConnectionName       string        `json:"connectionName"`
	CspSshKeyId          string        `json:"cspSshKeyId"`
	CspSshKeyName        string        `json:"cspSshKeyName"`
	Description          string        `json:"description"`
	Fingerprint          string        `json:"fingerprint"`
	Id                   string        `json:"id"`
	IsAutoGenerated      bool          `json:"isAutoGenerated"`
	KeyValueList         []KeyValueRes `json:"keyValueList"`
	Name                 string        `json:"name"`
	PrivateKey           string        `json:"privateKey"`
	PublicKey            string        `json:"publicKey"`
	SystemLabel          string        `json:"systemLabel"`
	Username             string        `json:"username"`
	VerifiedUsername     string        `json:"verifiedUsername"`
}

type SecureShellResList []SecureShellRes

type ImageRes struct {
	AssociatedObjectList []string      `json:"associatedObjectList"`
	ConnectionName       string        `json:"connectionName"`
	CreationDate         string        `json:"creationDate"`
	CspImageId           string        `json:"cspImageId"`
	CspImageName         string        `json:"cspImageName"`
	Description          string        `json:"description"`
	GuestOS              string        `json:"guestOS"`
	Id                   string        `json:"id"`
	IsAutoGenerated      bool          `json:"isAutoGenerated"`
	KeyValueList         []KeyValueRes `json:"keyValueList"`
	Name                 string        `json:"name"`
	Namespace            string        `json:"namespace"`
	Status               string        `json:"status"`
	SystemLabel          string        `json:"systemLabel"`
}

type SpecRes struct {
	AssociatedObjectList  []string `json:"associatedObjectList"`
	ConnectionName        string   `json:"connectionName"`
	CostPerHour           int      `json:"costPerHour"`
	CspSpecName           string   `json:"cspSpecName"`
	Description           string   `json:"description"`
	EbsBwMbps             int      `json:"ebsBwMbps"`
	EvaluationScore01     int      `json:"evaluationScore01"`
	EvaluationScore02     int      `json:"evaluationScore02"`
	EvaluationScore03     int      `json:"evaluationScore03"`
	EvaluationScore04     int      `json:"evaluationScore04"`
	EvaluationScore05     int      `json:"evaluationScore05"`
	EvaluationScore06     int      `json:"evaluationScore06"`
	EvaluationScore07     int      `json:"evaluationScore07"`
	EvaluationScore08     int      `json:"evaluationScore08"`
	EvaluationScore09     int      `json:"evaluationScore09"`
	EvaluationScore10     int      `json:"evaluationScore10"`
	EvaluationStatus      string   `json:"evaluationStatus"`
	GpuMemGiB             int      `json:"gpuMemGiB"`
	GpuModel              string   `json:"gpuModel"`
	GpuP2P                string   `json:"gpuP2p"`
	Id                    string   `json:"id"`
	IsAutoGenerated       bool     `json:"isAutoGenerated"`
	MaxNumStorage         int      `json:"maxNumStorage"`
	MaxTotalStorageTiB    int      `json:"maxTotalStorageTiB"`
	MemGiB                int      `json:"memGiB"`
	Name                  string   `json:"name"`
	Namespace             string   `json:"namespace"`
	NetBwGbps             int      `json:"netBwGbps"`
	NumCore               int      `json:"numCore"`
	NumGpu                int      `json:"numGpu"`
	NumStorage            int      `json:"numStorage"`
	NumvCPU               int      `json:"numvCPU"`
	OrderInFilteredResult int      `json:"orderInFilteredResult"`
	OsType                string   `json:"osType"`
	ProviderName          string   `json:"providerName"`
	RegionName            string   `json:"regionName"`
	RootDiskSize          string   `json:"rootDiskSize"`
	RootDiskType          string   `json:"rootDiskType"`
	StorageGiB            int      `json:"storageGiB"`
	SystemLabel           string   `json:"systemLabel"`
}

type RecommendVmResList []RecommendVmRes

type RecommendVmRes struct {
	Namespace             string  `json:"namespace"`
	ID                    string  `json:"id"`
	Name                  string  `json:"name"`
	ConnectionName        string  `json:"connectionName"`
	ProviderName          string  `json:"providerName"`
	RegionName            string  `json:"regionName"`
	CspSpecName           string  `json:"cspSpecName"`
	VCPU                  int     `json:"vCPU"`
	MemoryGiB             int     `json:"memoryGiB"`
	CostPerHour           float64 `json:"costPerHour"`
	OrderInFilteredResult int     `json:"orderInFilteredResult"`
	EvaluationScore01     float64 `json:"evaluationScore01"`
	EvaluationScore02     int     `json:"evaluationScore02"`
	EvaluationScore03     int     `json:"evaluationScore03"`
	EvaluationScore04     int     `json:"evaluationScore04"`
	EvaluationScore05     int     `json:"evaluationScore05"`
	EvaluationScore06     int     `json:"evaluationScore06"`
	EvaluationScore07     int     `json:"evaluationScore07"`
	EvaluationScore08     int     `json:"evaluationScore08"`
	EvaluationScore09     int     `json:"evaluationScore09"`
	EvaluationScore10     float64 `json:"evaluationScore10"`
	RootDiskType          string  `json:"rootDiskType"`
	RootDiskSize          string  `json:"rootDiskSize"`
}

type GetNsRes struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
