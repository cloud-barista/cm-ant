package tumblebug

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func (t *TumblebugClient) GetMcisIdsWithContext(ctx context.Context, nsId, mcisId string) ([]string, error) {
	url := t.withUrl(fmt.Sprintf("/ns/%s/mcis/%s?option=id", nsId, mcisId))
	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		log.Printf("error sending get mcis request: %v\n", err)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	var res struct {
		Output []string `json:"output"`
	}

	err = json.Unmarshal(resBytes, &res)

	if err != nil {
		log.Printf("error unmarshaling response body: %v\n", err)
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return res.Output, nil
}
func (t *TumblebugClient) GetMcisWithContext(ctx context.Context, nsId, mcisId string) (McisRes, error) {
	url := t.withUrl(fmt.Sprintf("/ns/%s/mcis/%s", nsId, mcisId))
	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		log.Printf("error sending get mcis request: %v\n", err)
		return McisRes{}, fmt.Errorf("failed to send request: %w", err)
	}

	mcisObject := McisRes{}

	err = json.Unmarshal(resBytes, &mcisObject)

	if err != nil {
		log.Printf("error unmarshaling response body: %v\n", err)
		return McisRes{}, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return mcisObject, nil
}

func (t *TumblebugClient) CommandToMcisWithContext(ctx context.Context, nsId, mcisId string, body SendCommandReq) (string, error) {

	url := t.withUrl(fmt.Sprintf("/ns/%s/cmd/mcis/%s", nsId, mcisId))

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		log.Println("send command request error", err)
		return "", err
	}

	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)

	if err != nil {
		log.Printf("error sending get mcis request: %v\n", err)
		return "", fmt.Errorf("failed to send request: %w", err)
	}

	ret := string(resBytes)

	return ret, nil
}
