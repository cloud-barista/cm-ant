package tumblebug

type SendCommandReq struct {
	Command  []string `json:"command"`
	UserName string   `json:"userName"`
}

type CreateNamespaceReq struct {
	Description string `json:"description"`
	Name        string `json:"name"`
}

type McisDynamicReq struct {
	Description     string              `json:"description"`
	InstallMonAgent string              `json:"installMonAgent"`
	Label           string              `json:"label"`
	Name            string              `json:"name"`
	SystemLabel     string              `json:"systemLabel"`
	VM              []VirtualMachineReq `json:"vm"`
}

type VirtualMachineReq struct {
	CommonImage    string `json:"commonImage"`
	CommonSpec     string `json:"commonSpec"`
	ConnectionName string `json:"connectionName"`
	Description    string `json:"description"`
	Label          string `json:"label"`
	Name           string `json:"name"`
	RootDiskSize   string `json:"rootDiskSize"`
	RootDiskType   string `json:"rootDiskType"`
	SubGroupSize   string `json:"subGroupSize"`
	VMUserPassword string `json:"vmUserPassword"`
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

type McisReq struct {
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
	Filter   Filter   `json:"filter"`
	Limit    string   `json:"limit"`
	Priority Priority `json:"priority"`
}
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
type Parameter struct {
	Key string   `json:"key"`
	Val []string `json:"val"`
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
	CommonImage    string `json:"commonImage"`
	CommonSpec     string `json:"commonSpec"`
	ConnectionName string `json:"connectionName"`
	Description    string `json:"description"`
	Label          string `json:"label"`
	Name           string `json:"name"`
	RootDiskSize   string `json:"rootDiskSize"`
	RootDiskType   string `json:"rootDiskType"`
	SubGroupSize   string `json:"subGroupSize"`
	VMUserPassword string `json:"vmUserPassword"`
}

type DynamicMcisReq struct {
	Description     string         `json:"description"`
	InstallMonAgent string         `json:"installMonAgent"`
	Label           string         `json:"label"`
	Name            string         `json:"name"`
	SystemLabel     string         `json:"systemLabel"`
	VM              []DynamicVmReq `json:"vm"`
}
