package outbound

import (
	"bytes"
	"fmt"
	"net/http"
)

func request(httpMethod, requestUrl string, authHeader string, body []byte) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(httpMethod, requestUrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("error while creating new request; %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if authHeader != "" {
		req.Header.Add("Authorization", authHeader)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error while request to requestUrl: %s; %w", requestUrl, err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("internal server error while request to requestUrl: %s; %w", requestUrl, err)
	}

	return resp, err
}
