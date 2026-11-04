package spider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	neturl "net/url"

	"github.com/rs/zerolog/log"
)

func (s *SpiderClient) GetPriceInfoWithContext(ctx context.Context, pf, regionName string, body PriceInfoReq) (CloudPriceDataRes, error) {

	var cloudPriceData CloudPriceDataRes

	// cb-spider v0.12.3 BREAKING: POST + FilterList(body) -> GET + query params
	//   - FilterList(body) is not sent. Client-side filtering in cm-ant is enough
	//     (price_collector.go: instanceType matching, priceValidator, price source selection)
	//   - ConnectionName is a required query parameter (v0.12.17 PriceInfoRest.go:72~ ConnectionRequest binding)
	//   - The route is fixed to GET /priceinfo/vm/{RegionName} (productFamily is hardcoded to "vm",
	//     cb-spider v0.12.17 PriceInfoRest.go:98 `cres.RSTypeString(cres.VM)`).
	//     The pf argument is still sent in the query for compatibility, but the server ignores it.
	q := neturl.Values{}
	q.Set("ConnectionName", body.ConnectionName)
	if pf != "" {
		// Ignored by the server; kept in the query as a debugging trace until a /priceinfo/{pf}/{region} route exists.
		q.Set("productFamily", pf)
		log.Info().Msgf("productFamily query (server-side ignored as of v0.12.17): %s", pf)
	}
	endpoint := fmt.Sprintf("/priceinfo/vm/%s?%s", regionName, q.Encode())
	url := s.withUrl(endpoint)

	log.Info().Msgf("CB-Spider API URL: %s", url)

	// cb-spider v0.12.6+ requires authentication; requestWithBaseAuthWithContext attaches the Authorization header.
	// (requestWithContext sends no auth header and gets 401.)
	resBytes, err := s.requestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)

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

	// cb-spider v0.12.6+ requires authentication for anycall as well
	resBytes, err := s.requestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)

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
