package tumblebug

type SendCommandReq struct {
	Command  []string `json:"command"`
	UserName string   `json:"userName"`
}

// MciCmdReq is struct for remote command (updated for latest cb-tumblebug)
type MciCmdReq struct {
	UserName string   `json:"userName" example:"cb-user" default:""`
	Command  []string `json:"command" validate:"required" example:"client_ip=$(echo $SSH_CLIENT | awk '{print $1}'); echo SSH client IP is: $client_ip"`
}

type CreateNamespaceReq struct {
	Description string `json:"description"`
	Name        string `json:"name"`
}

type MciDynamicReq struct {
	Description     string              `json:"description"`
	InstallMonAgent string              `json:"installMonAgent"`
	Label           map[string]string   `json:"label"` // v0.11.8: string -> map[string]string
	Name            string              `json:"name"`
	SystemLabel     string              `json:"systemLabel"`
	SubGroups       []VirtualMachineReq `json:"subGroups"` // v0.11.8: VM -> SubGroups
}

type VirtualMachineReq struct {
	ImageId        string            `json:"imageId"` // v0.11.8: commonImage -> imageId
	SpecId         string            `json:"specId"`  // v0.11.8: commonSpec -> specId
	ConnectionName string            `json:"connectionName"`
	Description    string            `json:"description"`
	Label          map[string]string `json:"label"` // v0.11.8: string -> map[string]string
	Name           string            `json:"name"`
	RootDiskSize   string            `json:"rootDiskSize"`
	RootDiskType   string            `json:"rootDiskType"`
	SubGroupSize   string            `json:"subGroupSize"`
	VMUserPassword string            `json:"vmUserPassword"`
}

type SecurityGroupReq struct {
	Name           string            `json:"name"`
	ConnectionName string            `json:"connectionName"`
	VNetID         string            `json:"vNetId"`
	Description    string            `json:"description"`
	FirewallRules  []FirewallRuleReq `json:"firewallRules"`
}

type FirewallRuleReq struct {
	FromPort   string `json:"fromPort"`
	ToPort     string `json:"toPort"`
	IPProtocol string `json:"ipprotocol"`
	Direction  string `json:"direction"`
	CIDR       string `json:"cidr"`
}

type SecureShellReq struct {
	ConnectionName string `json:"connectionName"`
	Name           string `json:"name"`
	Username       string `json:"username"`
	Description    string `json:"description"`
}

type ImageReq struct {
	ConnectionName string `json:"connectionName"`
	Name           string `json:"name"`
	CspImageId     string `json:"cspImageId"`
	Description    string `json:"description"`
	GuestOS        string `json:"guestOS"`
}

type SpecReq struct {
	ConnectionName string `json:"connectionName"`
	Name           string `json:"name"`
	CspSpecName    string `json:"cspSpecName"`
	Description    string `json:"description"`
}

type MciReq struct {
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	InstallMonAgent string  `json:"installMonAgent"`
	Label           string  `json:"label"`
	SystemLabel     string  `json:"systemLabel"`
	Vm              []VmReq `json:"vm"`
}

type VmReq struct {
	SubGroupSize     string   `json:"subGroupSize"`
	Name             string   `json:"name"`
	ImageId          string   `json:"imageId"`
	VmUserAccount    string   `json:"vmUserAccount"`
	ConnectionName   string   `json:"connectionName"`
	SshKeyId         string   `json:"sshKeyId"`
	SpecId           string   `json:"specId"`
	SecurityGroupIds []string `json:"securityGroupIds"`
	VNetId           string   `json:"vNetId"`
	SubnetId         string   `json:"subnetId"`
	Description      string   `json:"description"`
	VmUserPassword   string   `json:"vmUserPassword"`
	RootDiskType     string   `json:"rootDiskType"`
	RootDiskSize     string   `json:"rootDiskSize"`
}

type RecommendVmReq struct {
	Filter   FilterInfo   `json:"filter"`
	Limit    string       `json:"limit"`
	Priority PriorityInfo `json:"priority"`
}

// CB-Tumblebug v0.11.9+ 구조체들
type FilterInfo struct {
	Policy []FilterCondition `json:"policy"`
}

type FilterCondition struct {
	Metric    string      `json:"metric"`
	Condition []Operation `json:"condition"`
}

type Operation struct {
	Operator string `json:"operator"` // >=, <=, ==
	Operand  string `json:"operand"`
}

type PriorityInfo struct {
	Policy []PriorityCondition `json:"policy"`
}

type PriorityCondition struct {
	Metric    string      `json:"metric"`
	Parameter []Parameter `json:"parameter"`
}

type Parameter struct {
	Key string   `json:"key"`
	Val []string `json:"val"`
}

// 하위 호환성을 위한 기존 구조체들 (deprecated)
type Condition struct {
	Operand  string `json:"operand"`
	Operator string `json:"operator"`
}
type FilterPolicy struct {
	Condition []Condition `json:"condition"`
	Metric    string      `json:"metric"`
}
type Filter struct {
	Policy []FilterPolicy `json:"policy"`
}
type Policy struct {
	Metric    string      `json:"metric"`
	Parameter []Parameter `json:"parameter"`
}
type Priority struct {
	Policy []Policy `json:"policy"`
}
type CreateNsReq struct {
	Description string `json:"dscription"`
	Name        string `json:"name"`
}

type DynamicVmReq struct {
	ImageId        string            `json:"imageId"` // v0.11.8: commonImage -> imageId
	SpecId         string            `json:"specId"`  // v0.11.8: commonSpec -> specId
	ConnectionName string            `json:"connectionName"`
	Description    string            `json:"description"`
	Label          map[string]string `json:"label"`
	Name           string            `json:"name"`
	RootDiskSize   string            `json:"rootDiskSize"`
	RootDiskType   string            `json:"rootDiskType"`
	SubGroupSize   string            `json:"subGroupSize"`
	VMUserPassword string            `json:"vmUserPassword"`
	SshKeyId       string            `json:"sshKeyId,omitempty"` // SSH key for VM access
}

type DynamicMciReq struct {
	Description     string            `json:"description"`
	InstallMonAgent string            `json:"installMonAgent"`
	Label           map[string]string `json:"label"`
	Name            string            `json:"name"`
	SystemLabel     string            `json:"systemLabel"`
	SubGroups       []DynamicVmReq    `json:"subGroups"` // v0.11.8: VM -> SubGroups
}

// SSH Key related structures
type SshKeyReq struct {
	Name           string `json:"name" validate:"required"`
	ConnectionName string `json:"connectionName" validate:"required"`
	Description    string `json:"description"`
}

type SshKeyInfo struct {
	ResourceType     string `json:"resourceType"`
	Id               string `json:"id"`
	Uid              string `json:"uid,omitempty"`
	Name             string `json:"name"`
	ConnectionName   string `json:"connectionName,omitempty"`
	Description      string `json:"description,omitempty"`
	Username         string `json:"username,omitempty"`
	VerifiedUsername string `json:"verifiedUsername,omitempty"`
	PublicKey        string `json:"publicKey,omitempty"`
	PrivateKey       string `json:"privateKey,omitempty"`
}

// CB-Tumblebug v0.12.1 스마트 매칭 구조체
type SearchImageRequest struct {
	// MatchedSpecId is the ID of the matched spec.
	// If specified, only the images that match this spec will be returned.
	MatchedSpecId string `json:"matchedSpecId,omitempty"`

	// Cloud Service Provider (ex: "aws", "azure", "gcp", etc.)
	ProviderName string `json:"providerName,omitempty"`

	// Cloud Service Provider Region (ex: "us-east-1", "us-west-2", etc.)
	RegionName string `json:"regionName,omitempty"`

	// Simplified OS name and version string. Space-separated for AND condition (ex: "ubuntu 22.04")
	OSType string `json:"osType,omitempty"`

	// The architecture of the operating system of the image. (ex: "x86_64", "arm64", etc.)
	OSArchitecture string `json:"osArchitecture,omitempty"`

	// Whether the image is ready for GPU usage or not.
	IsGPUImage *bool `json:"isGPUImage,omitempty"`

	// Whether the image is specialized image only for Kubernetes nodes.
	IsKubernetesImage *bool `json:"isKubernetesImage,omitempty"`

	// Whether the image is registered by CB-Tumblebug asset file or not.
	IsRegisteredByAsset *bool `json:"isRegisteredByAsset,omitempty"`

	// Whether the search results should include deprecated images or not.
	IncludeDeprecatedImage *bool `json:"includeDeprecatedImage,omitempty"`

	// Return basic OS distribution only without additional applications.
	IncludeBasicImageOnly *bool `json:"includeBasicImageOnly,omitempty"`

	// Maximum number of images to be returned in the search results.
	MaxResults int `json:"maxResults,omitempty"`

	// Keywords for searching images in detail.
	DetailSearchKeys []string `json:"detailSearchKeys,omitempty"`
}

type SearchImageResponse struct {
	ImageCount int         `json:"imageCount"`
	ImageList  []ImageInfo `json:"imageList"`
}
