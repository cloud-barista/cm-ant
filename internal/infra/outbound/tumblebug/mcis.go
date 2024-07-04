package tumblebug

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/cloud-barista/cm-ant/pkg/utils"
)

func (t *TumblebugClient) GetMcisIdsWithContext(ctx context.Context, nsId, mcisId string) ([]string, error) {
	var res struct {
		Output []string `json:"output"`
	}

	url := t.withUrl(fmt.Sprintf("/ns/%s/mcis/%s?option=id", nsId, mcisId))
	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		utils.LogError("error sending get mcis id request:", err)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	err = json.Unmarshal(resBytes, &res)

	if err != nil {
		utils.LogError("error unmarshaling response body:", err)
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return res.Output, nil
}

func (t *TumblebugClient) GetMcisWithContext(ctx context.Context, nsId, mcisId string) (McisRes, error) {
	var mcisObject McisRes

	url := t.withUrl(fmt.Sprintf("/ns/%s/mcis/%s", nsId, mcisId))
	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		utils.LogError("error sending get mcis request:", err)

		if errors.Is(err, ErrInternalServerError) {
			return mcisObject, ErrNotFound
		}
		return mcisObject, fmt.Errorf("failed to send request: %w", err)
	}

	err = json.Unmarshal(resBytes, &mcisObject)

	if err != nil {
		utils.LogError("error unmarshaling response body:", err)
		return mcisObject, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return mcisObject, nil
}

func (t *TumblebugClient) CommandToMcisWithContext(ctx context.Context, nsId, mcisId string, body SendCommandReq) (string, error) {

	url := t.withUrl(fmt.Sprintf("/ns/%s/cmd/mcis/%s", nsId, mcisId))

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		utils.LogError("error marshaling request body:", err)
		return "", err
	}

	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)

	if err != nil {
		utils.LogError("error sending command to mcis request:", err)
		return "", fmt.Errorf("failed to send request: %w", err)
	}

	ret := string(resBytes)

	return ret, nil
}

func (t *TumblebugClient) CommandToVmWithContext(ctx context.Context, nsId, mcisId, vmId string, body SendCommandReq) (string, error) {

	url := t.withUrl(fmt.Sprintf("/ns/%s/cmd/mcis/%s?vmId=%s", nsId, mcisId, vmId))

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		utils.LogError("error marshaling request body:", err)
		return "", err
	}

	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)

	if err != nil {
		utils.LogError("error sending command to vm request:", err)
		return "", fmt.Errorf("failed to send request: %w", err)
	}

	ret := string(resBytes)

	return ret, nil
}

func (t *TumblebugClient) GetNsWithContext(ctx context.Context, nsId string) (GetNsRes, error) {
	var nsRes GetNsRes

	url := t.withUrl(fmt.Sprintf("/ns/%s", nsId))
	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		utils.LogError("error sending get mcis request:", err)
		return nsRes, fmt.Errorf("failed to send request: %w", err)
	}

	err = json.Unmarshal(resBytes, &nsRes)

	if err != nil {
		utils.LogError("error unmarshaling response body:", err)
		return nsRes, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return nsRes, nil
}

func (t *TumblebugClient) GetRecommendVmWithContext(ctx context.Context, body RecommendVmReq) (RecommendVmResList, error) {
	var res RecommendVmResList
	url := t.withUrl("/mcisRecommendVm")

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		utils.LogError("error marshaling request body:", err)
		return res, err
	}

	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)

	if err != nil {
		utils.LogError("error sending get recommend vm request:", err)
		return res, fmt.Errorf("failed to send request: %w", err)
	}

	err = json.Unmarshal(resBytes, &res)

	if err != nil {
		utils.LogError("error unmarshaling response body:", err)
		return res, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return res, nil
}

func (t *TumblebugClient) CreateNsWithContext(ctx context.Context, body CreateNsReq) error {
	url := t.withUrl("/ns")

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		utils.LogError("error marshaling request body:", err)
		return err
	}

	_, err = t.requestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)

	if err != nil {
		utils.LogError("error sending create ns request:", err)
		return fmt.Errorf("failed to send request: %w", err)
	}

	return nil
}

func (t *TumblebugClient) DynamicVmWithContext(ctx context.Context, nsId, mcisId string, body DynamicVmReq) (McisRes, error) {
	var res McisRes
	url := t.withUrl(fmt.Sprintf("/ns/%s/mcis/%s/vmDynamic", nsId, mcisId))

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		utils.LogError("error marshaling request body:", err)
		return res, err
	}

	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)

	if err != nil {
		utils.LogError("error sending dynamic vm request:", err)
		return res, fmt.Errorf("failed to send request: %w", err)
	}

	err = json.Unmarshal(resBytes, &res)

	if err != nil {
		utils.LogError("error unmarshaling response body:", err)
		return res, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return res, nil
}

func (t *TumblebugClient) DynamicMcisWithContext(ctx context.Context, nsId string, body DynamicMcisReq) (McisRes, error) {
	var res McisRes
	url := t.withUrl(fmt.Sprintf("/ns/%s/mcisDynamic", nsId))

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		utils.LogError("error marshaling request body:", err)
		return res, err
	}

	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)

	if err != nil {
		utils.LogError("error sending dynamic mcis request:", err)
		return res, fmt.Errorf("failed to send request: %w", err)
	}

	err = json.Unmarshal(resBytes, &res)

	if err != nil {
		utils.LogError("error unmarshaling response body:", err)
		return res, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return res, nil
}

// ControlLifecycleWithContext call tumblebug's control lifecycle api with specific action.
// action should be on of terminate | suspend | resume | reboot | refine | continue | withdraw
func (t *TumblebugClient) ControlLifecycleWithContext(ctx context.Context, nsId, mcisId, action string) error {
	url := t.withUrl(fmt.Sprintf("/ns/%s/control/mcis/%s?action=%s", nsId, mcisId, action))

	_, err := t.requestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		utils.LogError("error sending control lifecycle request:", err)
		return fmt.Errorf("failed to send request: %w", err)
	}

	return nil
}

// DeleteAllMcisWithContext call tumblebug's api which delete all mcis in ns.
// This should call after all the vm's in mcis is the status of terminate or suspend.
// If you want to change mcis's vm lifecycle use ControlLifecycleWithContext.
func (t *TumblebugClient) DeleteAllMcisWithContext(ctx context.Context, nsId string) error {
	url := t.withUrl(fmt.Sprintf("/ns/%s/mcis", nsId))

	_, err := t.requestWithBaseAuthWithContext(ctx, http.MethodDelete, url, nil)

	if err != nil {
		utils.LogError("error sending delete all mcis request:", err)
		return fmt.Errorf("failed to send request: %w", err)
	}

	return nil
}

// DeleteAllResourcesWithContext call tumblebug's api which delete all default resources in ns.
// This should call after DeleteAllMcisWithContext executed.
func (t *TumblebugClient) DeleteAllResourcesWithContext(ctx context.Context, nsId string) error {
	url := t.withUrl(fmt.Sprintf("/ns/%s/defaultResources", nsId))

	_, err := t.requestWithBaseAuthWithContext(ctx, http.MethodDelete, url, nil)

	if err != nil {
		utils.LogError("error sending delete all mcis request:", err)
		return fmt.Errorf("failed to send request: %w", err)
	}

	return nil
}
