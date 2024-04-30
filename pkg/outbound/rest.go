package outbound

import (
	"bytes"
	"fmt"
	"net/http"
)

func request(method, url string, body []byte, header map[string]string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("error while creating new request; %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	for k, v := range header {
		req.Header.Add(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error while request to url: %s; %w", url, err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("internal server error while request to url: %s; %w", url, err)
	}

	return resp, err
}

func RequestWithBaseAuth(method, url string, body []byte) (*http.Response, error) {
	return request(method, url, body, map[string]string{"Authorization": TumblebugBaseAuthHeader()})
}
