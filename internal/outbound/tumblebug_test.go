package outbound

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHostCheck(t *testing.T) {
	expected := "http://localhost:1323"
	t.Run("domain name check", func(t *testing.T) {
		got := HostWithPort()
		if got != expected {
			t.Errorf("got : %s, expected : %s", got, expected)
		}
	})
}

func TestSendCommandTo(t *testing.T) {
	type temp struct {
		Command  []string `json:"command"`
		UserName string   `json:"userName"`
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if (*r).Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "unauthenticated error !!!!!")
			return
		}

		decoder := json.NewDecoder(r.Body)
		requestBody := temp{}
		err := decoder.Decode(&requestBody)

		if err != nil {
			t.Errorf("error while parsing")
		}

		log.Printf("%+v\n", requestBody)
		w.WriteHeader(http.StatusOK)

		_, err = fmt.Fprintln(w, fmt.Sprintf("{\"message\": \"success\", \"endpoint\":\"%s\"}", (*r).URL.Path))
		if err != nil {
			return
		}
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("send command to tumblebug server", func(t *testing.T) {
		got, err := SendCommandTo(server.URL, "namespaceId", "mcisId", struct {
			Command  []string `json:"command"`
			UserName string   `json:"userName"`
		}{
			Command: []string{
				"command1",
				"command2",
			},
			UserName: "UserName - sehyeong",
		})
		if err != nil {
			t.Errorf("error created %v", err)
		}

		path := "/tumblebug/ns/namespaceId/cmd/mcis/mcisId"
		if !strings.Contains(got, path) {
			t.Errorf("error created")
		}
	})
}
