package spider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/cloud-barista/cm-ant/internal/utils"
)

const (
	productFamily = "ComputeInstance"
)

func (s *SpiderClient) GetPriceInfoWithContext(ctx context.Context, regionName string, body PriceInfoReq) (CloudPriceDataRes, error) {

	var cloudPriceData CloudPriceDataRes

	url := s.withUrl(fmt.Sprintf("/priceinfo/%s/%s", productFamily, regionName))

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		log.Println("marshalling body error;", err)
		return cloudPriceData, err
	}

	resBytes, err := s.requestWithContext(ctx, http.MethodPost, url, marshalledBody, nil)

	if err != nil {
		utils.LogError("error sending get price info data request:", err)

		return cloudPriceData, fmt.Errorf("failed to send request: %w", err)
	}

	err = json.Unmarshal(resBytes, &cloudPriceData)

	if err != nil {
		utils.LogError("error unmarshaling response body:", err)
		return cloudPriceData, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return cloudPriceData, nil
}

var (
	ErrSpiderCostResultEmpty = errors.New("cost information does not exist")
)

func (s *SpiderClient) GetCostWithResourceWithContext(ctx context.Context, body AnycallReq) (*CostWithResourcesRes, error) {
	var anycallRes AnycallRes
	var costWithResourceRes CostWithResourcesRes

	url := s.withUrl("/anycall")

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		log.Println("marshalling body error;", err)
		return &costWithResourceRes, err
	}

	resBytes, err := s.requestWithContext(ctx, http.MethodPost, url, marshalledBody, nil)

	if err != nil {
		utils.LogError("error sending get cost info data request:", err)

		return &costWithResourceRes, fmt.Errorf("failed to send request: %w", err)
	}

	err = json.Unmarshal(resBytes, &anycallRes)

	if err != nil {
		utils.LogError("error unmarshaling response body:", err)
		return &costWithResourceRes, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	if anycallRes.OKeyValueList == nil || len(anycallRes.IKeyValueList) == 0 {
		return &costWithResourceRes, ErrSpiderCostResultEmpty
	}

	err = json.Unmarshal([]byte(anycallRes.OKeyValueList[0].Value), &costWithResourceRes)

	if err != nil {
		utils.LogError("error unmarshaling :", err)
		return &costWithResourceRes, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return &costWithResourceRes, nil
}
