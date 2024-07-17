package spider

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cloud-barista/cm-ant/pkg/config"
	"github.com/cloud-barista/cm-ant/pkg/utils"
)

var (
	ErrBadRequest          = errors.New("bad request")
	ErrNotFound            = errors.New("object not found")
	ErrInternalServerError = errors.New("spider server has got error")
)

type SpiderClient struct {
	client     *http.Client
	host       string
	port       string
	username   string
	password   string
	domain     string
	authHeader string
}

func NewSpiderClient(client *http.Client) *SpiderClient {
	t := config.AppConfig.Spider

	var authHeader string
	if t.Username != "" && t.Password != "" {
		authHeader = fmt.Sprintf(
			"Basic %s",
			base64.StdEncoding.EncodeToString([]byte(
				fmt.Sprintf("%s:%s", t.Username, t.Password),
			),
			),
		)
	}
	return &SpiderClient{
		client:     client,
		host:       t.Host,
		port:       t.Port,
		username:   t.Username,
		password:   t.Password,
		domain:     fmt.Sprintf("%s:%s", t.Host, t.Port),
		authHeader: authHeader,
	}
}

func (t *SpiderClient) withUrl(endpoint string) string {
	trimmedEndpoint := strings.TrimPrefix(endpoint, "/")
	return fmt.Sprintf("%s/spider/%s", t.domain, trimmedEndpoint)
}

func (t *SpiderClient) requestWithContext(ctx context.Context, method, url string, body []byte, header map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		utils.LogErrorf("Failed to create request with context: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	for k, v := range header {
		req.Header.Add(k, v)
	}

	utils.LogInfof("Sending request to client with endpoint [%s - %s]\n", method, url)
	resp, err := t.client.Do(req)
	if err != nil {
		utils.LogErrorf("Failed to send request: %v", err)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		rb, _ := io.ReadAll(resp.Body)
		utils.LogErrorf("Unexpected status code: %d, response: %s", resp.StatusCode, string(rb))

		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrNotFound
		} else if resp.StatusCode == http.StatusInternalServerError {
			return nil, ErrInternalServerError
		} else if resp.StatusCode == http.StatusBadRequest {
			return nil, ErrBadRequest
		}

		return nil, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(rb))
	}

	rb, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.LogErrorf("Failed to read response body: %v", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	utils.LogInfo("Request with context completed successfully.")
	return rb, nil
}

func (t *SpiderClient) requestWithBaseAuthWithContext(ctx context.Context, method, url string, body []byte) ([]byte, error) {
	return t.requestWithContext(ctx, method, url, body, map[string]string{"Authorization": t.authHeader})
}
