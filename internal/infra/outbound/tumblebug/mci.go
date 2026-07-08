package tumblebug

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

// looksLikeNotFound reports whether a cb-tumblebug 500 error body indicates a genuinely
// missing resource. cb-tumblebug returns HTTP 500 (not 404) for absent infra/node, so
// GetMci/GetVm remap 500 to ErrNotFound only when the body says so — a transient 500 is
// propagated instead of being misread as "not found" and triggering a spurious recreate
// (BAR-1412 / FR-MA2-PERF-007-09 hardening).
func looksLikeNotFound(err error) bool {
	if err == nil {
		return false
	}
	m := strings.ToLower(err.Error())
	return strings.Contains(m, "does not exist") ||
		strings.Contains(m, "no node found") ||
		strings.Contains(m, "not exist") ||
		strings.Contains(m, "not found")
}

func (t *TumblebugClient) GetMciWithContext(ctx context.Context, nsId, mciId string) (MciRes, error) {
	var mciObject MciRes

	// cb-tumblebug v0.12.7 BREAKING: /mci/ → /infra/
	url := t.withUrl(fmt.Sprintf("/ns/%s/infra/%s", nsId, mciId))
	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		log.Error().Msgf("error sending get mci request; %v", err)

		if errors.Is(err, ErrInternalServerError) && looksLikeNotFound(err) {
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

// GetVmWithContext retrieves specific VM information
func (t *TumblebugClient) GetVmWithContext(ctx context.Context, nsId, mciId, vmId string) (VmInfo, error) {
	var vmObject VmInfo

	// cb-tumblebug v0.12.7 BREAKING: /mci/ → /infra/, /vm/ → /node/
	url := t.withUrl(fmt.Sprintf("/ns/%s/infra/%s/node/%s", nsId, mciId, vmId))
	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		log.Error().Msgf("error sending get vm request; %v", err)

		if errors.Is(err, ErrInternalServerError) && looksLikeNotFound(err) {
			return vmObject, ErrNotFound
		}
		return vmObject, fmt.Errorf("failed to send request: %w", err)
	}

	err = json.Unmarshal(resBytes, &vmObject)

	if err != nil {
		log.Error().Msgf("error unmarshaling response body; %v", err)
		return vmObject, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return vmObject, nil
}

// GetAvailableImagesWithContext retrieves available images for a specific connection
func (t *TumblebugClient) GetAvailableImagesWithContext(ctx context.Context, connectionName string) ([]ImageInfo, error) {
	var response struct {
		Image []ImageInfo `json:"image"`
	}

	url := t.withUrl(fmt.Sprintf("/ns/%s/resources/image?connectionName=%s", "system", connectionName))
	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		log.Error().Msgf("error sending get available images request; %v", err)
		return response.Image, fmt.Errorf("failed to send request: %w", err)
	}

	err = json.Unmarshal(resBytes, &response)
	if err != nil {
		log.Error().Msgf("error unmarshaling response body; %v", err)
		return response.Image, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return response.Image, nil
}

// SearchImagesWithContext searches images using CB-Tumblebug v0.11.8+ smart matching
func (t *TumblebugClient) SearchImagesWithContext(ctx context.Context, nsId string, req SearchImageRequest) ([]ImageInfo, error) {
	var response SearchImageResponse

	url := t.withUrl(fmt.Sprintf("/ns/%s/resources/searchImage", nsId))
	marshalledBody, err := json.Marshal(req)
	if err != nil {
		log.Error().Msgf("error marshaling search image request body; %v", err)
		return response.ImageList, err
	}

	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)
	if err != nil {
		log.Error().Msgf("error sending search image request; %v", err)
		return response.ImageList, fmt.Errorf("failed to send request: %w", err)
	}

	err = json.Unmarshal(resBytes, &response)
	if err != nil {
		log.Error().Msgf("error unmarshaling search image response body; %v", err)
		return response.ImageList, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	log.Info().Msgf("SearchImages found %d images for criteria: provider=%s, region=%s, osType=%s, osArchitecture=%s",
		response.ImageCount, req.ProviderName, req.RegionName, req.OSType, req.OSArchitecture)

	return response.ImageList, nil
}

// CreateSshKeyWithContext creates an SSH key in CB-Tumblebug
func (t *TumblebugClient) CreateSshKeyWithContext(ctx context.Context, nsId string, sshKeyReq SshKeyReq) (SshKeyInfo, error) {
	var sshKeyInfo SshKeyInfo

	url := t.withUrl(fmt.Sprintf("/ns/%s/resources/sshKey", nsId))
	marshalledBody, err := json.Marshal(sshKeyReq)
	if err != nil {
		log.Error().Msgf("error marshaling ssh key request body; %v", err)
		return sshKeyInfo, err
	}

	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)
	if err != nil {
		log.Error().Msgf("error sending create ssh key request; %v", err)
		return sshKeyInfo, fmt.Errorf("failed to send request: %w", err)
	}

	err = json.Unmarshal(resBytes, &sshKeyInfo)
	if err != nil {
		log.Error().Msgf("error unmarshaling ssh key response body; %v", err)
		return sshKeyInfo, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return sshKeyInfo, nil
}

// GetSshKeyWithContext retrieves an SSH key from CB-Tumblebug
func (t *TumblebugClient) GetSshKeyWithContext(ctx context.Context, nsId, sshKeyId string) (SshKeyInfo, error) {
	var sshKeyInfo SshKeyInfo

	url := t.withUrl(fmt.Sprintf("/ns/%s/resources/sshKey/%s", nsId, sshKeyId))
	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Error().Msgf("error sending get ssh key request; %v", err)
		return sshKeyInfo, fmt.Errorf("failed to send request: %w", err)
	}

	err = json.Unmarshal(resBytes, &sshKeyInfo)
	if err != nil {
		log.Error().Msgf("error unmarshaling ssh key response body; %v", err)
		return sshKeyInfo, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return sshKeyInfo, nil
}

func (t *TumblebugClient) CommandToMciWithContext(ctx context.Context, nsId, mciId string, body SendCommandReq) (string, error) {

	// cb-tumblebug v0.12.7 BREAKING: /cmd/mci/ → /cmd/infra/
	url := t.withUrl(fmt.Sprintf("/ns/%s/cmd/infra/%s", nsId, mciId))

	// Convert SendCommandReq to MciCmdReq for latest cb-tumblebug compatibility
	mciCmdReq := MciCmdReq{
		UserName: body.UserName,
		Command:  body.Command,
	}

	marshalledBody, err := json.Marshal(mciCmdReq)
	if err != nil {
		log.Error().Msgf("error marshaling request body; %v", err)
		return "", err
	}

	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)

	if err != nil {
		log.Error().Msgf("error sending command to mci request; %v", err)
		return "", fmt.Errorf("failed to send request: %w", err)
	}

	// Parse the response to extract error information if any
	var result MciSshCmdResultForAPI
	if err := json.Unmarshal(resBytes, &result); err != nil {
		log.Error().Msgf("error unmarshaling response body; %v", err)
		// Return the raw response if parsing fails
		return string(resBytes), nil
	}

	// Check if there are any errors in the results
	for _, res := range result.Results {
		if res.Error != "" {
			log.Error().Msgf("command execution error for VM %s: %s", res.NodeId, res.Error)
			return "", fmt.Errorf("command execution failed for VM %s: %s", res.NodeId, res.Error)
		}
	}

	// Return the marshaled response for backward compatibility
	return string(resBytes), nil
}

func (t *TumblebugClient) CommandToVmWithContext(ctx context.Context, nsId, mciId, vmId string, body SendCommandReq) (string, error) {

	// cb-tumblebug v0.12.7 BREAKING: /cmd/mci/ → /cmd/infra/, vmId → nodeId
	url := t.withUrl(fmt.Sprintf("/ns/%s/cmd/infra/%s?nodeId=%s", nsId, mciId, vmId))

	// Convert SendCommandReq to MciCmdReq for latest cb-tumblebug compatibility
	mciCmdReq := MciCmdReq{
		UserName: body.UserName,
		Command:  body.Command,
	}

	marshalledBody, err := json.Marshal(mciCmdReq)
	if err != nil {
		log.Error().Msgf("error marshaling request body; %v", err)
		return "", err
	}

	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)

	if err != nil {
		log.Error().Msgf("error sending command to vm request; %v", err)
		return "", fmt.Errorf("failed to send request: %w", err)
	}

	// Parse the response to extract error information if any
	var result MciSshCmdResultForAPI
	if err := json.Unmarshal(resBytes, &result); err != nil {
		log.Error().Msgf("error unmarshaling response body; %v", err)
		// Return the raw response if parsing fails
		return string(resBytes), nil
	}

	// Check if there are any errors in the results
	for _, res := range result.Results {
		if res.Error != "" {
			log.Error().Msgf("command execution error for VM %s: %s", res.NodeId, res.Error)
			return "", fmt.Errorf("command execution failed for VM %s: %s", res.NodeId, res.Error)
		}
	}

	// Return the marshaled response for backward compatibility
	return string(resBytes), nil
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

func (t *TumblebugClient) GetRecommendVmWithContext(ctx context.Context, body RecommendVmReq) (SpecInfoList, error) {
	var res SpecInfoList
	// CB-Tumblebug v0.11.9에서 엔드포인트가 변경됨: /mciRecommendVm -> /recommendSpec
	url := t.withUrl("/recommendSpec")

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
	// cb-tumblebug v0.12.7 BREAKING: vmDynamic → nodeGroupDynamic, /mci/ → /infra/
	// (v0.12.9 server.go:412 — RestPostInfraNodeGroupDynamic)
	url := t.withUrl(fmt.Sprintf("/ns/%s/infra/%s/nodeGroupDynamic", nsId, mciId))

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
	// cb-tumblebug v0.12.7 BREAKING: mciDynamic → infraDynamic
	url := t.withUrl(fmt.Sprintf("/ns/%s/infraDynamic", nsId))

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
	// cb-tumblebug v0.12.7 BREAKING: /control/mci/ → /control/infra/
	url := t.withUrl(fmt.Sprintf("/ns/%s/control/infra/%s?action=%s", nsId, mciId, action))

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
//
// option is passed as the tumblebug `option` query parameter. Use "terminate" or
// "force" to purge an MCI whose backing CSP resources are already gone (both return
// 200 even when the VM no longer exists); pass "" for the default behaviour.
func (t *TumblebugClient) DeleteAllMciWithContext(ctx context.Context, nsId, option string) error {
	// cb-tumblebug v0.12.7 BREAKING: /mci → /infra
	url := t.withUrl(fmt.Sprintf("/ns/%s/infra", nsId))
	if option != "" {
		url = fmt.Sprintf("%s?option=%s", url, option)
	}

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
