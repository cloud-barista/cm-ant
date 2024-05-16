package spider

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
)

func requestWitContext(ctx context.Context, method, url string, body []byte, header map[string]string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		log.Println("error while creating new request with context")
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	for k, v := range header {
		req.Header.Add(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("error while request to client", err)
		return resp, err
	}

	return resp, nil
}


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

	return resp, err
}
