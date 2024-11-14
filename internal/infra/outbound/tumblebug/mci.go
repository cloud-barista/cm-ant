package tumblebug

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

func (t *TumblebugClient) GetMciWithContext(ctx context.Context, nsId, mciId string) (MciRes, error) {
	var mciObject MciRes

	url := t.withUrl(fmt.Sprintf("/ns/%s/mci/%s", nsId, mciId))
	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		log.Error().Msgf("error sending get mci request; %v", err)

		if errors.Is(err, ErrInternalServerError) {
			return mciObject, ErrNotFound
		}
		return mciObject, fmt.Errorf("failed to send request: %w", err)
	}

	err = json.Unmarshal(resBytes, &mciObject)

	if err != nil {
		log.Error().Msgf("error unmarshaling response body; %v", err)
		return mciObject, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return mciObject, nil
}

func (t *TumblebugClient) CommandToMciWithContext(ctx context.Context, nsId, mciId string, body SendCommandReq) (string, error) {

	url := t.withUrl(fmt.Sprintf("/ns/%s/cmd/mci/%s", nsId, mciId))

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		log.Error().Msgf("error marshaling request body; %v", err)
		return "", err
	}

	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)

	if err != nil {
		log.Error().Msgf("error sending command to mci request; %v", err)
		return "", fmt.Errorf("failed to send request: %w", err)
	}

	ret := string(resBytes)

	return ret, nil
}

func (t *TumblebugClient) CommandToVmWithContext(ctx context.Context, nsId, mciId, vmId string, body SendCommandReq) (string, error) {

	url := t.withUrl(fmt.Sprintf("/ns/%s/cmd/mci/%s?vmId=%s", nsId, mciId, vmId))

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		log.Error().Msgf("error marshaling request body; %v", err)
		return "", err
	}

	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)

	if err != nil {
		log.Error().Msgf("error sending command to vm request; %v", err)
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
		log.Error().Msgf("error sending get mci request; %v", err)
		return nsRes, fmt.Errorf("failed to send request: %w", err)
	}

	err = json.Unmarshal(resBytes, &nsRes)

	if err != nil {
		log.Error().Msgf("error unmarshaling response body; %v", err)
		return nsRes, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return nsRes, nil
}

func (t *TumblebugClient) GetRecommendVmWithContext(ctx context.Context, body RecommendVmReq) (RecommendVmResList, error) {
	var res RecommendVmResList
	url := t.withUrl("/mciRecommendVm")

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		log.Error().Msgf("error marshaling request body; %v", err)
		return res, err
	}

	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)

	if err != nil {
		log.Error().Msgf("error sending get recommend vm request; %v", err)
		return res, fmt.Errorf("failed to send request: %w", err)
	}

	err = json.Unmarshal(resBytes, &res)

	if err != nil {
		log.Error().Msgf("error unmarshaling response body; %v", err)
		return res, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return res, nil
}

func (t *TumblebugClient) CreateNsWithContext(ctx context.Context, body CreateNsReq) error {
	url := t.withUrl("/ns")

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		log.Error().Msgf("error marshaling request body; %v", err)
		return err
	}

	_, err = t.requestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)

	if err != nil {
		log.Error().Msgf("error sending create ns request; %v", err)
		return fmt.Errorf("failed to send request: %w", err)
	}

	return nil
}

func (t *TumblebugClient) DynamicVmWithContext(ctx context.Context, nsId, mciId string, body DynamicVmReq) (MciRes, error) {
	var res MciRes
	url := t.withUrl(fmt.Sprintf("/ns/%s/mci/%s/vmDynamic", nsId, mciId))

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		log.Error().Msgf("error marshaling request body; %v", err)
		return res, err
	}

	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)

	if err != nil {
		log.Error().Msgf("error sending dynamic vm request; %v", err)
		return res, fmt.Errorf("failed to send request: %w", err)
	}

	err = json.Unmarshal(resBytes, &res)

	if err != nil {
		log.Error().Msgf("error unmarshaling response body; %v", err)
		return res, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return res, nil
}

func (t *TumblebugClient) DynamicMciWithContext(ctx context.Context, nsId string, body DynamicMciReq) (MciRes, error) {
	var res MciRes
	url := t.withUrl(fmt.Sprintf("/ns/%s/mciDynamic", nsId))

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		log.Error().Msgf("error marshaling request body; %v", err)
		return res, err
	}

	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)

	if err != nil {
		log.Error().Msgf("error sending dynamic mci request; %v", err)
		return res, fmt.Errorf("failed to send request: %w", err)
	}

	err = json.Unmarshal(resBytes, &res)

	if err != nil {
		log.Error().Msgf("error unmarshaling response body; %v", err)
		return res, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return res, nil
}

// ControlLifecycleWithContext call tumblebug's control lifecycle api with specific action.
// action should be on of terminate | suspend | resume | reboot | refine | continue | withdraw
func (t *TumblebugClient) ControlLifecycleWithContext(ctx context.Context, nsId, mciId, action string) error {
	url := t.withUrl(fmt.Sprintf("/ns/%s/control/mci/%s?action=%s", nsId, mciId, action))

	_, err := t.requestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		log.Error().Msgf("error sending control lifecycle request; %v", err)
		return fmt.Errorf("failed to send request: %w", err)
	}

	return nil
}

// DeleteAllMciWithContext call tumblebug's api which delete all mci in ns.
// This should call after all the vm's in mci is the status of terminate or suspend.
// If you want to change mci's vm lifecycle use ControlLifecycleWithContext.
func (t *TumblebugClient) DeleteAllMciWithContext(ctx context.Context, nsId string) error {
	url := t.withUrl(fmt.Sprintf("/ns/%s/mci", nsId))

	_, err := t.requestWithBaseAuthWithContext(ctx, http.MethodDelete, url, nil)

	if err != nil {
		log.Error().Msgf("error sending delete all mci request; %v", err)
		return fmt.Errorf("failed to send request: %w", err)
	}

	return nil
}

// DeleteAllResourcesWithContext call tumblebug's api which delete all default resources in ns.
// This should call after DeleteAllMciWithContext executed.
func (t *TumblebugClient) DeleteAllResourcesWithContext(ctx context.Context, nsId string) error {
	url := t.withUrl(fmt.Sprintf("/ns/%s/defaultResources", nsId))

	_, err := t.requestWithBaseAuthWithContext(ctx, http.MethodDelete, url, nil)

	if err != nil {
		log.Error().Msgf("error sending delete all mci request; %v", err)
		return fmt.Errorf("failed to send request: %w", err)
	}

	return nil
}
