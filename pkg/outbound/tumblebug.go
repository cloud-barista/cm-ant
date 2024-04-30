package outbound

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

func SendCommandTo(nsId, mcisId string, body SendCommandReq) (string, error) {
	tumblebugUrl := TumblebugHostWithPort()
	url := fmt.Sprintf("%s/tumblebug/ns/%s/cmd/mcis/%s", tumblebugUrl, nsId, mcisId)

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		log.Println("send command request error", err)
		return "", err
	}

	res, err := RequestWithBaseAuth(http.MethodPost, url, marshalledBody)

	if err != nil {
		log.Println("send command request error", errors.Unwrap(err))
		return "", err
	}

	responseBody, err := io.ReadAll(res.Body)

	if err != nil {
		log.Println("send command request error", errors.Unwrap(err))
		return "", err
	}
	defer res.Body.Close()

	ret := string(responseBody)
	log.Println(ret)

	return ret, nil
}

func MockMigrate(createNamespaceBody CreateNamespaceReq, mcisDynamicBody McisDynamicReq) error {
	tumblebugUrl := TumblebugHostWithPort()
	// ns create
	url := fmt.Sprintf("%s/tumblebug/ns", tumblebugUrl)

	marshalledBody, err := json.Marshal(createNamespaceBody)
	if err != nil {
		log.Println("mock migrate request error", err)
		return err
	}
	res, err := RequestWithBaseAuth(http.MethodPost, url, marshalledBody)

	if err != nil {
		log.Println("mock migrate request error", errors.Unwrap(err))
		return err
	}

	responseBody, err := io.ReadAll(res.Body)

	if err != nil {
		log.Println("mock migrate request error", errors.Unwrap(err))
		return err
	}
	defer res.Body.Close()

	var result map[string]interface{}
	err = json.Unmarshal(responseBody, &result)
	if err != nil {
		return err
	}

	nsId, ok := result["id"]
	if !ok {
		return fmt.Errorf("ns is not created correctly")
	}

	log.Println("ns created! ", nsId)

	url = fmt.Sprintf("%s/tumblebug/ns/%s/mcisDynamic", tumblebugUrl, nsId)
	marshalledBody, err = json.Marshal(mcisDynamicBody)
	if err != nil {
		log.Println("mock migrate request error1", err)
		return err
	}

	log.Println("request url is ", url)
	res, err = RequestWithBaseAuth(http.MethodPost, url, marshalledBody)

	if err != nil {
		log.Println("mock migrate request error2", errors.Unwrap(err))
		responseBody, _ = io.ReadAll(res.Body)
		log.Println(string(responseBody))
		return err
	}

	responseBody, err = io.ReadAll(res.Body)

	if err != nil {
		log.Println("mock migrate request error3", errors.Unwrap(err))
		log.Println(string(responseBody))
		return err
	}
	defer res.Body.Close()

	err = json.Unmarshal(responseBody, &result)
	if err != nil {
		return err
	}
	log.Println(result)

	return nil
}
