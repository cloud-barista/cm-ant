package spider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

var spiderUrl = SpiderHostWithPort()

func GetProductFamilyWitContext(ctx context.Context, regionName, connectionName string) (ProductfamilyRes, error) {
	url := fmt.Sprintf("%s/spider/productfamily/%s?ConnectionName=%s", spiderUrl, regionName, connectionName)
	var productfamily ProductfamilyRes
	res, err := requestWitContext(ctx, http.MethodGet, url, nil, nil)
	if err != nil {
		log.Println("get mcis object request error: ", err)
		return productfamily, err
	}

	rb, err := io.ReadAll(res.Body)

	if err != nil {
		log.Println("read response body error:", err)
		return productfamily, err
	}
	defer res.Body.Close()

	err = json.Unmarshal(rb, &productfamily)

	if err != nil {
		return productfamily, err
	}

	return productfamily, nil
}

func GetPriceInfoWithContext(ctx context.Context, productfamily, regionName string, body PriceInfoReq) (CloudPriceDataRes, error) {
	url := fmt.Sprintf("%s/spider/priceinfo/%s/%s", spiderUrl, productfamily, regionName)

	var cloudPriceData CloudPriceDataRes
	marshalledBody, err := json.Marshal(body)
	if err != nil {
		log.Println("marshalling body error;", err)
		return cloudPriceData, err
	}

	res, err := requestWitContext(ctx, http.MethodPost, url, marshalledBody, nil)

	if err != nil {
		log.Println("send command request error;", errors.Unwrap(err))
		return cloudPriceData, err
	}

	responseBody, err := io.ReadAll(res.Body)

	if err != nil {
		log.Println("read response body error;", err)
		return cloudPriceData, err
	}
	defer res.Body.Close()

	err = json.Unmarshal(responseBody, &cloudPriceData)

	if err != nil {
		return cloudPriceData, err
	}

	return cloudPriceData, nil
}
