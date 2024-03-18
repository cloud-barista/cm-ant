package outbound

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloud-barista/cm-ant/internal/common/configuration"
	om "github.com/cloud-barista/cm-ant/internal/outbound/models"
	"io"
	"log"
	"net/http"
)

func HostWithPort() string {
	config := configuration.Get().Tumblebug
	return fmt.Sprintf("%s:%s", config.URL, config.Port)
}

func baseAuthHeader() string {
	config := configuration.Get().Tumblebug
	header := fmt.Sprintf("%s:%s", config.Username, config.Password)
	encodedHeader := base64.StdEncoding.EncodeToString([]byte(header))
	return fmt.Sprintf("Basic %s", encodedHeader)
}

func requestWithBaseAuth(method, url string, body []byte) (*http.Response, error) {
	return request(method, url, baseAuthHeader(), body)
}

func SendCommandTo(domain, namespaceId, mcisId string, body om.SendCommandRequestBody) (string, error) {
	url := fmt.Sprintf("%s/tumblebug/ns/%s/cmd/mcis/%s", domain, namespaceId, mcisId)

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		log.Println("send command request error", err)
		return "", err
	}

	res, err := requestWithBaseAuth(http.MethodPost, url, marshalledBody)

	if err != nil {
		log.Println("send command request error", errors.Unwrap(err))
		return "", err
	}

	responseBody, err := io.ReadAll(res.Body)

	if err != nil {
		log.Println("send command request error", errors.Unwrap(err))
		return "", err
	}
	defer res.Body.Close()

	return string(responseBody), nil
}
