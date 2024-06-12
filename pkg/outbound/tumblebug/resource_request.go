package tumblebug

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

var ResourcesNotFound = errors.New("resources not found")

func CreateSpecWithContext(ctx context.Context, nsId string, specReq SpecReq) (SpecRes, error) {
	url := fmt.Sprintf("%s/tumblebug/ns/%s/resources/spec", tumblebugUrl, nsId)
	var spec SpecRes
	marshalledBody, err := json.Marshal(specReq)
	if err != nil {
		log.Println("marshalling body error;", err)
		return spec, err
	}

	res, err := RequestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)
	if err != nil {
		log.Println("create image request error: ", err)
		return spec, err
	}

	rb, err := io.ReadAll(res.Body)

	if err != nil {
		log.Println("read response body error:", err)
		return spec, err
	}
	defer res.Body.Close()

	err = json.Unmarshal(rb, &spec)

	if err != nil {
		return spec, err
	}

	return spec, nil
}

func CreateImageWithContext(ctx context.Context, nsId string, imageReq ImageReq) (ImageRes, error) {
	url := fmt.Sprintf("%s/tumblebug/ns/%s/resources/image?action=registerWithId", tumblebugUrl, nsId)
	var image ImageRes
	marshalledBody, err := json.Marshal(imageReq)
	if err != nil {
		log.Println("marshalling body error;", err)
		return image, err
	}

	res, err := RequestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)
	if err != nil {
		log.Println("create image request error: ", err)
		return image, err
	}

	rb, err := io.ReadAll(res.Body)

	if err != nil {
		log.Println("read response body error:", err)
		return image, err
	}
	defer res.Body.Close()

	err = json.Unmarshal(rb, &image)

	if err != nil {
		return image, err
	}

	return image, nil
}

func CreateSecureShellWithContext(ctx context.Context, nsId string, securityGroupReq SecureShellReq) (SecureShellRes, error) {
	url := fmt.Sprintf("%s/tumblebug/ns/%s/resources/sshKey", tumblebugUrl, nsId)
	var ssh SecureShellRes
	marshalledBody, err := json.Marshal(securityGroupReq)
	if err != nil {
		log.Println("marshalling body error;", err)
		return ssh, err
	}

	res, err := RequestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)
	if err != nil {
		log.Println("create secure shell request error: ", err)
		return ssh, err
	}

	rb, err := io.ReadAll(res.Body)

	if err != nil {
		log.Println("read response body error:", err)
		return ssh, err
	}
	defer res.Body.Close()

	err = json.Unmarshal(rb, &ssh)

	if err != nil {
		return ssh, err
	}

	return ssh, nil
}

func CreateSecurityGroupWithContext(ctx context.Context, nsId string, securityGroupReq SecurityGroupReq) (SecurityGroupRes, error) {
	url := fmt.Sprintf("%s/tumblebug/ns/%s/resources/securityGroup", tumblebugUrl, nsId)
	var sg SecurityGroupRes
	marshalledBody, err := json.Marshal(securityGroupReq)
	if err != nil {
		log.Println("marshalling body error;", err)
		return sg, err
	}

	res, err := RequestWithBaseAuthWithContext(ctx, http.MethodPost, url, marshalledBody)
	if err != nil {
		log.Println("create security group request error: ", err)
		return sg, err
	}

	rb, err := io.ReadAll(res.Body)

	if err != nil {
		log.Println("read response body error:", err)
		return sg, err
	}
	defer res.Body.Close()

	err = json.Unmarshal(rb, &sg)

	if err != nil {
		return sg, err
	}

	return sg, nil
}

func GetMcisWithContext(ctx context.Context, nsId, mcisId string) (McisRes, error) {
	url := fmt.Sprintf("%s/tumblebug/ns/%s/mcis/%s", tumblebugUrl, nsId, mcisId)
	var mcis McisRes
	res, err := RequestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Println("get spec request error: ", err)
		return mcis, err
	}

	rb, err := io.ReadAll(res.Body)

	if err != nil {
		log.Println("read response body error:", err)
		return mcis, err
	}
	defer res.Body.Close()
	stringBody := string(rb)
	if res.StatusCode == http.StatusInternalServerError && strings.Contains(stringBody, "not exist") {
		return mcis, ResourcesNotFound
	}

	err = json.Unmarshal(rb, &mcis)

	if err != nil {
		return mcis, err
	}

	return mcis, nil
}

func GetSpecWithContext(ctx context.Context, nsId string, specId string) error {
	url := fmt.Sprintf("%s/tumblebug/ns/%s/resources/spec/%s", tumblebugUrl, nsId, specId)

	res, err := RequestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Println("get spec request error: ", err)
		return err
	}

	rb, err := io.ReadAll(res.Body)

	if err != nil {
		log.Println("read response body error:", err)
		return err
	}
	defer res.Body.Close()
	stringBody := string(rb)
	if res.StatusCode == http.StatusBadRequest && strings.Contains(stringBody, "Failed to find") {
		return ResourcesNotFound
	}

	log.Println(stringBody)

	return nil
}

func GetImageWithContext(ctx context.Context, nsId string, imageId string) error {
	url := fmt.Sprintf("%s/tumblebug/ns/%s/resources/image/%s", tumblebugUrl, nsId, imageId)
	res, err := RequestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Println("get image request error: ", err)
		return err
	}

	rb, err := io.ReadAll(res.Body)

	if err != nil {
		log.Println("read response body error:", err)
		return err
	}
	defer res.Body.Close()
	stringBody := string(rb)
	if res.StatusCode == http.StatusBadRequest && strings.Contains(stringBody, "Failed to find") {
		return ResourcesNotFound
	}

	log.Println(stringBody)

	return nil
}

func GetSecureShellWithContext(ctx context.Context, nsId, sshId string) (SecureShellRes, error) {
	url := fmt.Sprintf("%s/tumblebug/ns/%s/resources/sshKey/%s", tumblebugUrl, nsId, sshId)
	var ssh SecureShellRes
	res, err := RequestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Println("get secure shell request error: ", err)
		return ssh, err
	}

	rb, err := io.ReadAll(res.Body)

	if err != nil {
		log.Println("read response body error:", err)
		return ssh, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusBadRequest && strings.Contains(string(rb), "Failed to find") {
		return ssh, ResourcesNotFound
	}

	err = json.Unmarshal(rb, &ssh)

	if err != nil {
		return ssh, err
	}

	return ssh, nil
}

func GetSecurityGroupWithContext(ctx context.Context, nsId, sgId string) error {
	url := fmt.Sprintf("%s/tumblebug/ns/%s/resources/securityGroup/%s", tumblebugUrl, nsId, sgId)
	res, err := RequestWithBaseAuthWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Println("get security group request error: ", err)
		return err
	}
	rb, err := io.ReadAll(res.Body)

	if err != nil {
		log.Println("read response body error:", err)
		return err
	}
	defer res.Body.Close()

	stringBody := string(rb)
	if res.StatusCode == http.StatusBadRequest && strings.Contains(stringBody, "Failed to find") {
		return ResourcesNotFound
	}

	log.Println(stringBody)

	return nil
}
