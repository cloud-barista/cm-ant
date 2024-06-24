package tumblebug

import "net/http"

type TumblebugClient struct {
	client *http.Client
}

func NewTumblebugClient(client *http.Client) *TumblebugClient {
	return &TumblebugClient{
		client: client,
	}
}
