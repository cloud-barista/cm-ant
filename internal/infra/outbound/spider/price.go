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

	// cb-spider v0.12.3 BREAKING: POST + FilterList(body) → GET + query params
	//   - FilterList(body) 미전송. cm-ant는 클라이언트 측 필터링으로 충분
	//     (price_collector.go: instanceType 매칭, priceValidator, CSPPriceInfo의 terms["OnDemand"] 추출)
	//   - ConnectionName은 query parameter 필수 (v0.12.17 PriceInfoRest.go:72~ ConnectionRequest 바인딩)
	//   - 라우트는 GET /priceinfo/vm/{RegionName} 으로 고정 (productFamily는 "vm" 하드코딩,
	//     cb-spider v0.12.17 PriceInfoRest.go:98 `cres.RSTypeString(cres.VM)`).
	//     cm-ant의 pf 인자는 호환을 위해 query에 그대로 보내지만 *서버는 무시* — 향후 정리 후보.
	q := neturl.Values{}
	q.Set("ConnectionName", body.ConnectionName)
	if pf != "" {
		// 서버 무시지만 디버깅 흔적용으로 query에 표기. 정식 productFamily 다중화는 별도 라우트(/priceinfo/{pf}/{region}) 도입 시 보정.
		q.Set("productFamily", pf)
		log.Info().Msgf("productFamily query (server-side ignored as of v0.12.17): %s", pf)
	}
	endpoint := fmt.Sprintf("/priceinfo/vm/%s?%s", regionName, q.Encode())
	url := s.withUrl(endpoint)

	log.Info().Msgf("CB-Spider API URL: %s", url)

	// cb-spider v0.12.6+ 강제 인증 — requestWithBaseAuthWithContext가 Authorization 헤더 자동 첨부.
	// (기존 requestWithContext는 인증 헤더 미송신 → 401 회귀 원인이었음, BAR-685)
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

	// cb-spider v0.12.6+ 강제 인증 — anycall도 Authorization 헤더 송신 필요(BAR-685)
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
