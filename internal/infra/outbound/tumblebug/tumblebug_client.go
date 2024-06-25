package tumblebug

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/cloud-barista/cm-ant/pkg/config"
	"github.com/cloud-barista/cm-ant/pkg/utils"
)

type TumblebugClient struct {
	client     *http.Client
	host       string
	port       string
	username   string
	password   string
	domain     string
	authHeader string
}

func NewTumblebugClient(client *http.Client) *TumblebugClient {
	t := config.AppConfig.Tumblebug
	return &TumblebugClient{
		client:   client,
		host:     t.Host,
		port:     t.Port,
		username: t.Username,
		password: t.Password,
		domain:   fmt.Sprintf("%s:%s", t.Host, t.Port),
		authHeader: fmt.Sprintf(
			"Basic %s",
			base64.StdEncoding.EncodeToString([]byte(
				fmt.Sprintf("%s:%s", t.Username, t.Password),
			),
			),
		),
	}
}

func (t *TumblebugClient) withUrl(endpoint string) string {
	trimmedEndpoint := strings.TrimPrefix(endpoint, "/")
	return fmt.Sprintf("%s/tumblebug/%s", t.domain, trimmedEndpoint)
}

func (t *TumblebugClient) requestWithContext(ctx context.Context, method, url string, body []byte, header map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("[ERROR] Failed to create request with context: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	for k, v := range header {
		req.Header.Add(k, v)
	}

	utils.LogInfof("Sending request to client with endpoint [%s - %s]\n", method, url)
	resp, err := t.client.Do(req)
	if err != nil {
		log.Printf("[ERROR] Failed to send request: %v", err)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		rb, _ := io.ReadAll(resp.Body)
		log.Printf("[ERROR] Unexpected status code: %d, response: %s", resp.StatusCode, string(rb))
		return nil, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(rb))
	}

	rb, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read response body: %v", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	utils.LogInfo("Request with context completed successfully.")
	return rb, nil
}

func (t *TumblebugClient) requestWithBaseAuthWithContext(ctx context.Context, method, url string, body []byte) ([]byte, error) {
	return t.requestWithContext(ctx, method, url, body, map[string]string{"Authorization": t.authHeader})
}
