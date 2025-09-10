package spider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

func (s *SpiderClient) GetPriceInfoWithContext(ctx context.Context, pf, regionName string, body PriceInfoReq) (CloudPriceDataRes, error) {

	var cloudPriceData CloudPriceDataRes

	// Change URL to match v0.11.5 API: /priceinfo/{productFamily}/{regionName} -> /priceinfo/vm/{regionName}
	url := s.withUrl(fmt.Sprintf("/priceinfo/vm/%s", regionName))

	// Add productFamily to FilterList (create copy to avoid modifying original body)
	requestBody := body
	if pf != "" {
		requestBody.FilterList = append(requestBody.FilterList, FilterReq{
			Key:   "productFamily",
			Value: pf,
		})
		log.Info().Msgf("Added productFamily to FilterList: %s", pf)
	}

	log.Info().Msgf("CB-Spider API URL: %s", url)
	log.Info().Msgf("Request body: %+v", requestBody)

	marshalledBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Error().Msgf("marshalling body error; %v", err)
		return cloudPriceData, err
	}

	resBytes, err := s.requestWithContext(ctx, http.MethodPost, url, marshalledBody, nil)

	if err != nil {
		log.Error().Msgf("error sending get price info data request; %v", err)

		return cloudPriceData, fmt.Errorf("failed to send request: %w", err)
	}

	err = json.Unmarshal(resBytes, &cloudPriceData)

	if err != nil {
		log.Error().Msgf("error unmarshaling response body; %v", err)
		return cloudPriceData, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	// Add logs for CB-Spider response debugging
	log.Info().Msgf("CB-Spider response length: %d bytes", len(resBytes))
	log.Info().Msgf("CB-Spider response: %s", string(resBytes))
	log.Info().Msgf("Parsed CloudPriceList count: %d", len(cloudPriceData.CloudPriceList))

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
		log.Error().Msgf("marshalling body error; %v", err)
		return &costWithResourceRes, err
	}

	resBytes, err := s.requestWithContext(ctx, http.MethodPost, url, marshalledBody, nil)

	if err != nil {
		log.Error().Msgf("error sending get cost info data request; %v", err)

		return &costWithResourceRes, fmt.Errorf("failed to send request: %w", err)
	}

	err = json.Unmarshal(resBytes, &anycallRes)

	if err != nil {
		log.Error().Msgf("error unmarshaling response body; %v", err)
		return &costWithResourceRes, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	if anycallRes.OKeyValueList == nil || len(anycallRes.IKeyValueList) == 0 {
		return &costWithResourceRes, ErrSpiderCostResultEmpty
	}

	err = json.Unmarshal([]byte(anycallRes.OKeyValueList[0].Value), &costWithResourceRes)

	if err != nil {
		log.Error().Msgf("error unmarshaling; %v", err)
		return &costWithResourceRes, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return &costWithResourceRes, nil
}
