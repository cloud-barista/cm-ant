package outbound

type SendCommandRequestBody struct {
	Command  []string `json:"command"`
	UserName string   `json:"userName"`
}
