package tumblebug

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cloud-barista/cm-ant/internal/utils"
)

func (t *TumblebugClient) GetSecureShellWithContext(ctx context.Context, nsId string) (SecureShellResList, error) {
	var res SecureShellResList
	url := t.withUrl(fmt.Sprintf("/ns/%s/resources/sshKey", nsId))

	resBytes, err := t.requestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		utils.LogError("error sending get secure shell request:", err)
		return res, fmt.Errorf("failed to send request: %w", err)
	}

	err = json.Unmarshal(resBytes, &res)

	if err != nil {
		utils.LogError("error unmarshaling response body:", err)
		return res, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return res, nil
}
