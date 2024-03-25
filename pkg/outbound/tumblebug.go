package outbound

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloud-barista/cm-ant/pkg/configuration"
	"io"
	"log"
	"net/http"
)

func RequestWithBaseAuth(method, url string, body []byte) (*http.Response, error) {
	return request(method, url, configuration.TumblebugBaseAuthHeader(), body)
}

func SendCommandTo(domain, nsId, mcisId string, body SendCommandReq) (string, error) {
	url := fmt.Sprintf("%s/tumblebug/ns/%s/cmd/mcis/%s", domain, nsId, mcisId)

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		log.Println("send command request error", err)
		return "", err
	}

	res, err := RequestWithBaseAuth(http.MethodPost, url, marshalledBody)

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

	ret := string(responseBody)
	log.Println(ret)

	return ret, nil
}
