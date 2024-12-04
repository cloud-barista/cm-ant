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

	url := s.withUrl(fmt.Sprintf("/priceinfo/%s/%s", pf, regionName))

	marshalledBody, err := json.Marshal(body)
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
