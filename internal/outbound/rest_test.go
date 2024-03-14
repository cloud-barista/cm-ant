package outbound

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hello/tumblebug" {
			t.Errorf("올바른 엔드포인트가 아닙니다. 기대값: /hello/tumblebug, 실제값: %s", r.URL.Path)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprint(w, "success")
		if err != nil {
			return
		}
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	response, err := request(http.MethodGet, fmt.Sprintf("%s/hello/tumblebug", server.URL), "", nil)
	if err != nil {
		t.Errorf("API 호출 실패: %v", err)
	}

	expectedBody := "success"
	expectedStatus := http.StatusOK

	got, err := io.ReadAll(response.Body)
	if err != nil {
		t.Errorf("API 호출 실패: %v", err)
	}

	if string(got) != expectedBody || response.StatusCode != expectedStatus {
		t.Errorf("올바른 응답이 아닙니다. 기대값: %s, %d, 실제값: %s, %d", expectedBody, expectedStatus, string(got), response.StatusCode)
	}
}
