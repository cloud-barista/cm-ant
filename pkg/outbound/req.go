package outbound

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
