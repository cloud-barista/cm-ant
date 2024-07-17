package spider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/cloud-barista/cm-ant/pkg/utils"
)

func (s *SpiderClient) GetPriceInfoWithContext(ctx context.Context, productfamily, regionName string, body PriceInfoReq) (CloudPriceDataRes, error) {

	var cloudPriceData CloudPriceDataRes

	url := s.withUrl(fmt.Sprintf("/priceinfo/%s/%s", productfamily, regionName))

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		log.Println("marshalling body error;", err)
		return cloudPriceData, err
	}

	resBytes, err := s.requestWithContext(ctx, http.MethodPost, url, marshalledBody, nil)

	if err != nil {
		utils.LogError("error sending get price info data request:", err)

		if errors.Is(err, ErrInternalServerError) {
			return cloudPriceData, ErrNotFound
		}
		return cloudPriceData, fmt.Errorf("failed to send request: %w", err)
	}

	err = json.Unmarshal(resBytes, &cloudPriceData)

	if err != nil {
		utils.LogError("error unmarshaling response body:", err)
		return cloudPriceData, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return cloudPriceData, nil
}
