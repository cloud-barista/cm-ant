package tumblebug

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

var tumblebugUrl = TumblebugHostWithPort()

func CreateMcisWithContext(ctx context.Context, nsId string, mcisReq McisReq) (McisRes, error) {
	url := fmt.Sprintf("%s/tumblebug/ns/%s/mcis", tumblebugUrl, nsId)
	var mcis McisRes
	marshalledBody, err := json.Marshal(mcisReq)
	if err != nil {
		log.Println("marshalling body error;", err)
		return mcis, err
	}

	res, err := RequestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)
	if err != nil {
		log.Println("create image request error: ", err)
		return mcis, err
	}

	rb, err := io.ReadAll(res.Body)

	if err != nil {
		log.Println("read response body error:", err)
		return mcis, err
	}
	defer res.Body.Close()

	err = json.Unmarshal(rb, &mcis)

	if err != nil {
		return mcis, err
	}

	return mcis, nil
}

func GetVmWithContext(ctx context.Context, nsId, mcisId, vmId string) (VmRes, error) {
	url := fmt.Sprintf("%s/tumblebug/ns/%s/mcis/%s/vm/%s", tumblebugUrl, nsId, mcisId, vmId)
	vm := VmRes{}

	res, err := RequestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Println("get mcis object request error: ", err)
		return vm, err
	}

	rb, err := io.ReadAll(res.Body)

	if err != nil {
		log.Println("read response body error:", err)
		return vm, err
	}
	defer res.Body.Close()

	err = json.Unmarshal(rb, &vm)

	if err != nil {
		return vm, err
	}

	return vm, nil
}

func GetMcisObjectWithContext(ctx context.Context, nsId, mcisId string) (McisRes, error) {
	url := fmt.Sprintf("%s/tumblebug/ns/%s/mcis/%s", tumblebugUrl, nsId, mcisId)

	res, err := RequestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Println("get mcis object request error: ", err)
		return McisRes{}, err
	}

	rb, err := io.ReadAll(res.Body)

	if err != nil {
		log.Println("read response body error:", err)
		return McisRes{}, err
	}
	defer res.Body.Close()

	mcisObject := McisRes{}

	err = json.Unmarshal(rb, &mcisObject)

	if err != nil {
		return McisRes{}, err
	}

	return mcisObject, nil
}

func CommandToVmWithContext(ctx context.Context, nsId, mcisId, vmId string, body SendCommandReq) (string, error) {
	url := fmt.Sprintf("%s/tumblebug/ns/%s/cmd/mcis/%s?vmId=%s", tumblebugUrl, nsId, mcisId, vmId)

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		log.Println("marshalling body error;", err)
		return "", err
	}

	res, err := RequestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)

	if err != nil {
		log.Println("send command request error;", errors.Unwrap(err))
		return "", err
	}

	responseBody, err := io.ReadAll(res.Body)

	if err != nil {
		log.Println("read response body error;", err)
		return "", err
	}
	defer res.Body.Close()

	ret := string(responseBody)
	log.Println("[command result]\n", ret)

	return ret, nil
}

func GetMcisObject(nsId, mcisId string) (McisRes, error) {
	url := fmt.Sprintf("%s/tumblebug/ns/%s/mcis/%s", tumblebugUrl, nsId, mcisId)

	res, err := RequestWithBaseAuth(http.MethodGet, url, nil)
	if err != nil {
		log.Println("get mcis object request error: ", err)
		return McisRes{}, err
	}

	rb, err := io.ReadAll(res.Body)

	if err != nil {
		log.Println("read response body error:", err)
		return McisRes{}, err
	}
	defer res.Body.Close()

	mcisObject := McisRes{}

	err = json.Unmarshal(rb, &mcisObject)

	if err != nil {
		return McisRes{}, err
	}

	return mcisObject, nil
}

func CommandToVm(nsId, mcisId, vmId string, body SendCommandReq) (string, error) {
	url := fmt.Sprintf("%s/tumblebug/ns/%s/cmd/mcis/%s?vmId=%s", tumblebugUrl, nsId, mcisId, vmId)

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
	log.Println("[command result]\n", ret)

	return ret, nil
}

func CommandToMcis(nsId, mcisId string, body SendCommandReq) (string, error) {
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
