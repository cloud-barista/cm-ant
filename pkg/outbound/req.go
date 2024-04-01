package outbound

type SendCommandReq struct {
	Command  []string `json:"command"`
	UserName string   `json:"userName"`
}
