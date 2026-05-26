package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// invokeRunLoadTest builds a JSON request and calls runLoadTest, returning the
// resulting *echo.HTTPError (nil if the handler returned no error).
//
// services is left nil on purpose: every assertion below stops at the request
// validation stage, which runs before any loadService call, so no real service
// or DB is required.
func invokeRunLoadTest(t *testing.T, body string) *echo.HTTPError {
	t.Helper()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/load/tests/run", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())

	s := &AntServer{}
	err := s.runLoadTest(c)
	if err == nil {
		return nil
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T: %v", err, err)
	}
	return he
}

func httpErrMessage(he *echo.HTTPError) string {
	if he == nil {
		return ""
	}
	if ar, ok := he.Message.(AntResponse[string]); ok {
		return ar.ErrorMessage
	}
	return ""
}

// TestRunLoadTest_ParamRangeValidation verifies the fixed range validation:
//   - out-of-range parameters must be rejected with HTTP 400 (this is the
//     regression that the `&&` bug previously let through),
//   - in-range parameters must pass range validation (verified indirectly by
//     stopping at the next check, "http request required", with empty httpReqs).
func TestRunLoadTest_ParamRangeValidation(t *testing.T) {
	const base = `"installLoadGenerator":{"installLocation":"local"}`

	cases := []struct {
		name      string
		body      string
		wantMsgIn string
	}{
		{
			name:      "VirtualUsers over range (100000) -> 400",
			body:      `{` + base + `,"virtualUsers":"100000","duration":"10","rampUpTime":"5","rampUpSteps":"2"}`,
			wantMsgIn: "virtual user",
		},
		{
			name:      "Duration over range (99999) -> 400",
			body:      `{` + base + `,"virtualUsers":"10","duration":"99999","rampUpTime":"5","rampUpSteps":"2"}`,
			wantMsgIn: "duration",
		},
		{
			name:      "RampUpTime over range (9999) -> 400",
			body:      `{` + base + `,"virtualUsers":"10","duration":"10","rampUpTime":"9999","rampUpSteps":"2"}`,
			wantMsgIn: "ramp up time",
		},
		{
			name:      "RampUpSteps over range (9999) -> 400",
			body:      `{` + base + `,"virtualUsers":"10","duration":"10","rampUpTime":"5","rampUpSteps":"9999"}`,
			wantMsgIn: "ramp up steps",
		},
		{
			name:      "Port over range (99999) -> 400",
			body:      `{` + base + `,"virtualUsers":"10","duration":"10","rampUpTime":"5","rampUpSteps":"2","httpReqs":[{"method":"GET","protocol":"http","hostname":"127.0.0.1","port":"99999","path":"/"}]}`,
			wantMsgIn: "Port",
		},
		{
			// In-range params must pass the range checks; with empty httpReqs the
			// handler then stops at the "http request required" check (still 400 but
			// a different message), proving the range validation did not reject them.
			name:      "in-range params pass range validation",
			body:      `{` + base + `,"virtualUsers":"100","duration":"300","rampUpTime":"60","rampUpSteps":"20","httpReqs":[]}`,
			wantMsgIn: "http request",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			he := invokeRunLoadTest(t, tc.body)
			if he == nil {
				t.Fatalf("expected HTTP 400, got nil (range validation passed unexpectedly)")
			}
			if he.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d (msg=%q)", he.Code, httpErrMessage(he))
			}
			if msg := httpErrMessage(he); !strings.Contains(msg, tc.wantMsgIn) {
				t.Fatalf("expected message to contain %q, got %q", tc.wantMsgIn, msg)
			}
		})
	}
}
