package tumblebug

import "time"

type Mcis struct {
	ID                            string      `json:"id,omitempty"`
	Name                          string      `json:"name,omitempty"`
	Status                        string      `json:"status,omitempty"`
	StatusCount                   StatusCount `json:"statusCount,omitempty"`
	TargetStatus                  string      `json:"targetStatus,omitempty"`
	TargetAction                  string      `json:"targetAction,omitempty"`
	InstallMonAgent               string      `json:"installMonAgent,omitempty"`
	ConfigureCloudAdaptiveNetwork string      `json:"configureCloudAdaptiveNetwork,omitempty"`
	Label                         string      `json:"label,omitempty"`
	SystemLabel                   string      `json:"systemLabel,omitempty"`
	SystemMessage                 string      `json:"systemMessage,omitempty"`
	Description                   string      `json:"description,omitempty"`
	VM                            []VM        `json:"vm,omitempty"`
	NewVMList                     any         `json:"newVmList,omitempty"`
}
type StatusCount struct {
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
type Location struct {
	Latitude     string `json:"latitude,omitempty"`
	Longitude    string `json:"longitude,omitempty"`
	BriefAddr    string `json:"briefAddr,omitempty"`
	CloudType    string `json:"cloudType,omitempty"`
	NativeRegion string `json:"nativeRegion,omitempty"`
}
type Region struct {
	Region string `json:"Region,omitempty"`
	Zone   string `json:"Zone,omitempty"`
}
type ConnectionConfig struct {
	ConfigName     string   `json:"ConfigName,omitempty"`
	ProviderName   string   `json:"ProviderName,omitempty"`
	DriverName     string   `json:"DriverName,omitempty"`
	CredentialName string   `json:"CredentialName,omitempty"`
	RegionName     string   `json:"RegionName,omitempty"`
	Location       Location `json:"Location,omitempty"`
}
type IID struct {
	NameID   string `json:"NameId,omitempty"`
	SystemID string `json:"SystemId,omitempty"`
}
type ImageIID struct {
	NameID   string `json:"NameId,omitempty"`
	SystemID string `json:"SystemId,omitempty"`
}
type VpcIID struct {
	NameID   string `json:"NameId,omitempty"`
	SystemID string `json:"SystemId,omitempty"`
}
type SubnetIID struct {
	NameID   string `json:"NameId,omitempty"`
	SystemID string `json:"SystemId,omitempty"`
}
type SecurityGroupIIds struct {
	NameID   string `json:"NameId,omitempty"`
	SystemID string `json:"SystemId,omitempty"`
}
type KeyPairIID struct {
	NameID   string `json:"NameId,omitempty"`
	SystemID string `json:"SystemId,omitempty"`
}
type KeyValueList struct {
	Key   string `json:"Key,omitempty"`
	Value string `json:"Value,omitempty"`
}
type CspViewVMDetail struct {
	Name               string              `json:"Name,omitempty"`
	ImageName          string              `json:"ImageName,omitempty"`
	VPCName            string              `json:"VPCName,omitempty"`
	SubnetName         string              `json:"SubnetName,omitempty"`
	SecurityGroupNames any                 `json:"SecurityGroupNames,omitempty"`
	KeyPairName        string              `json:"KeyPairName,omitempty"`
	CSPid              string              `json:"CSPid,omitempty"`
	DataDiskNames      any                 `json:"DataDiskNames,omitempty"`
	VMSpecName         string              `json:"VMSpecName,omitempty"`
	VMUserID           string              `json:"VMUserId,omitempty"`
	VMUserPasswd       string              `json:"VMUserPasswd,omitempty"`
	RootDiskType       string              `json:"RootDiskType,omitempty"`
	RootDiskSize       string              `json:"RootDiskSize,omitempty"`
	ImageType          string              `json:"ImageType,omitempty"`
	IID                IID                 `json:"IId,omitempty"`
	ImageIID           ImageIID            `json:"ImageIId,omitempty"`
	VpcIID             VpcIID              `json:"VpcIID,omitempty"`
	SubnetIID          SubnetIID           `json:"SubnetIID,omitempty"`
	SecurityGroupIIds  []SecurityGroupIIds `json:"SecurityGroupIIds,omitempty"`
	KeyPairIID         KeyPairIID          `json:"KeyPairIId,omitempty"`
	DataDiskIIDs       []string            `json:"DataDiskIIDs,omitempty"`
	StartTime          time.Time           `json:"StartTime,omitempty"`
	Region             Region              `json:"Region,omitempty"`
	NetworkInterface   string              `json:"NetworkInterface,omitempty"`
	PublicIP           string              `json:"PublicIP,omitempty"`
	PublicDNS          string              `json:"PublicDNS,omitempty"`
	PrivateIP          string              `json:"PrivateIP,omitempty"`
	PrivateDNS         string              `json:"PrivateDNS,omitempty"`
	RootDeviceName     string              `json:"RootDeviceName,omitempty"`
	SSHAccessPoint     string              `json:"SSHAccessPoint,omitempty"`
	KeyValueList       []KeyValueList      `json:"KeyValueList,omitempty"`
}
type VM struct {
	ID                 string           `json:"id,omitempty"`
	Name               string           `json:"name,omitempty"`
	IDByCSP            string           `json:"idByCSP,omitempty"`
	SubGroupID         string           `json:"subGroupId,omitempty"`
	Location           Location         `json:"location,omitempty"`
	Status             string           `json:"status,omitempty"`
	TargetStatus       string           `json:"targetStatus,omitempty"`
	TargetAction       string           `json:"targetAction,omitempty"`
	MonAgentStatus     string           `json:"monAgentStatus,omitempty"`
	NetworkAgentStatus string           `json:"networkAgentStatus,omitempty"`
	SystemMessage      string           `json:"systemMessage,omitempty"`
	CreatedTime        string           `json:"createdTime,omitempty"`
	Label              string           `json:"label,omitempty"`
	Description        string           `json:"description,omitempty"`
	Region             Region           `json:"region,omitempty"`
	PublicIP           string           `json:"publicIP,omitempty"`
	SSHPort            string           `json:"sshPort,omitempty"`
	PublicDNS          string           `json:"publicDNS,omitempty"`
	PrivateIP          string           `json:"privateIP,omitempty"`
	PrivateDNS         string           `json:"privateDNS,omitempty"`
	RootDiskType       string           `json:"rootDiskType,omitempty"`
	RootDiskSize       string           `json:"rootDiskSize,omitempty"`
	RootDeviceName     string           `json:"rootDeviceName,omitempty"`
	ConnectionName     string           `json:"connectionName,omitempty"`
	ConnectionConfig   ConnectionConfig `json:"connectionConfig,omitempty"`
	SpecID             string           `json:"specId,omitempty"`
	ImageID            string           `json:"imageId,omitempty"`
	VNetID             string           `json:"vNetId,omitempty"`
	SubnetID           string           `json:"subnetId,omitempty"`
	SecurityGroupIds   []string         `json:"securityGroupIds,omitempty"`
	DataDiskIds        []any            `json:"dataDiskIds,omitempty"`
	SSHKeyID           string           `json:"sshKeyId,omitempty"`
	VMUserAccount      string           `json:"vmUserAccount,omitempty"`
	CspViewVMDetail    CspViewVMDetail  `json:"cspViewVmDetail,omitempty"`
}
