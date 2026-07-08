package tumblebug

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cloud-barista/cm-ant/internal/config"
	"github.com/rs/zerolog/log"
)

var (
	ErrBadRequest          = errors.New("bad request")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrNotFound            = errors.New("object not found")
	ErrNotReady            = errors.New("cb-tumblebug not ready")
	ErrNotInitialized      = errors.New("cb-tumblebug not initialized")
	ErrInternalServerError = errors.New("tumblebug server has got error")
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

// readyzResponse mirrors the relevant fields of cb-tumblebug's ReadyzResponse
// (src/interface/rest/server/common/utility.go RestGetReadyz). cb-tumblebug
// returns HTTP 200 even when Initialized is false (per STANDARD-READYZ pattern
// B), so callers must inspect the body to determine actual readiness.
type readyzResponse struct {
	Ready       bool   `json:"ready"`
	Initialized bool   `json:"initialized"`
	Message     string `json:"message"`
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

func (t *TumblebugClient) Endpoint() string {
	return t.domain
}

func (t *TumblebugClient) withUrl(endpoint string) string {
	trimmedEndpoint := strings.TrimPrefix(endpoint, "/")
	return fmt.Sprintf("%s/tumblebug/%s", t.domain, trimmedEndpoint)
}

func (t *TumblebugClient) requestWithContext(ctx context.Context, method, url string, body []byte, header map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		log.Error().Msgf("failed to create request with context: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	for k, v := range header {
		req.Header.Add(k, v)
	}

	log.Info().Msgf("Sending request to client with endpoint [%s - %s]", method, url)
	resp, err := t.client.Do(req)
	if err != nil {
		log.Error().Msgf("failed to send request: %v", err)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		rb, _ := io.ReadAll(resp.Body)
		log.Error().Msgf("unexpected status code: %d, response: %s", resp.StatusCode, string(rb))

		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrNotFound
		} else if resp.StatusCode == http.StatusUnauthorized {
			return nil, ErrUnauthorized
		} else if resp.StatusCode == http.StatusInternalServerError {
			// Carry the response body so callers can tell a genuine not-found (cb-tumblebug
			// returns 500 for absent infra/node) from a transient server error (BAR-1412).
			return nil, fmt.Errorf("%w: %s", ErrInternalServerError, string(rb))
		} else if resp.StatusCode == http.StatusBadRequest {
			return nil, ErrBadRequest
		}

		return nil, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(rb))
	}

	rb, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Msgf("failed to read response body: %v", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	log.Info().Msg("Request with context completed successfully.")
	return rb, nil
}

func (t *TumblebugClient) requestWithBaseAuthWithContext(ctx context.Context, method, url string, body []byte) ([]byte, error) {
	return t.requestWithContext(ctx, method, url, body, map[string]string{"Authorization": t.authHeader})
}

// ReadyzWithContext calls cb-tumblebug GET /tumblebug/readyz and inspects the
// response body. Per STANDARD-READYZ §5, cb-tumblebug returns HTTP 200 even
// when SystemInitialized is false, so a simple status check is insufficient.
// We require both Ready and Initialized to be true.
func (t *TumblebugClient) ReadyzWithContext(ctx context.Context) error {
	url := t.withUrl("/readyz")
	body, err := t.requestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Error().Msgf("error sending cb-tumblebug readyz request; %v", err)
		return err
	}

	var rz readyzResponse
	if uerr := json.Unmarshal(body, &rz); uerr != nil {
		return fmt.Errorf("failed to parse cb-tumblebug readyz response: %w", uerr)
	}

	if !rz.Ready {
		return fmt.Errorf("%w: %s", ErrNotReady, rz.Message)
	}
	if !rz.Initialized {
		return fmt.Errorf("%w (Ready=true, Initialized=false): %s", ErrNotInitialized, rz.Message)
	}
	return nil
}

// AuthCheckWithContext calls a lightweight authenticated cb-tumblebug endpoint
// (GET /tumblebug/cloudInfo) to verify Basic Auth credentials. Per
// STANDARD-READYZ, /tumblebug/readyz is on the auth-middleware skip list and
// does not validate credentials. /tumblebug/cloudInfo is enforced and returns
// a small static cloud metadata payload that is independent of operator data.
func (t *TumblebugClient) AuthCheckWithContext(ctx context.Context) error {
	url := t.withUrl("/cloudInfo")
	_, err := t.requestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Error().Msgf("cb-tumblebug auth check failed (GET /tumblebug/cloudInfo); %v", err)
		return err
	}
	return nil
}
