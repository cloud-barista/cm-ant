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
//   - out-of-range values (above max AND below min) and non-numeric values
//     must be rejected with HTTP 400 (the `&&` bug previously let valid
//     out-of-range integers — both too-large and too-small, e.g. "0" — pass),
//   - in-range values (including min/max boundaries) must pass range validation
//     (verified by stopping at the next check, "http request required").
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

		// --- below range (v < 1): also broken by the && bug ("0" previously passed) ---
		{
			name:      "VirtualUsers below range (0) -> 400",
			body:      `{` + base + `,"virtualUsers":"0","duration":"10","rampUpTime":"5","rampUpSteps":"2"}`,
			wantMsgIn: "virtual user",
		},
		{
			name:      "Duration below range (0) -> 400",
			body:      `{` + base + `,"virtualUsers":"10","duration":"0","rampUpTime":"5","rampUpSteps":"2"}`,
			wantMsgIn: "duration",
		},
		{
			name:      "RampUpTime below range (0) -> 400",
			body:      `{` + base + `,"virtualUsers":"10","duration":"10","rampUpTime":"0","rampUpSteps":"2"}`,
			wantMsgIn: "ramp up time",
		},
		{
			name:      "RampUpSteps below range (0) -> 400",
			body:      `{` + base + `,"virtualUsers":"10","duration":"10","rampUpTime":"5","rampUpSteps":"0"}`,
			wantMsgIn: "ramp up steps",
		},
		{
			name:      "Port below range (0) -> 400",
			body:      `{` + base + `,"virtualUsers":"10","duration":"10","rampUpTime":"5","rampUpSteps":"2","httpReqs":[{"method":"GET","protocol":"http","hostname":"127.0.0.1","port":"0","path":"/"}]}`,
			wantMsgIn: "Port",
		},
		{
			name:      "VirtualUsers non-numeric (abc) -> 400",
			body:      `{` + base + `,"virtualUsers":"abc","duration":"10","rampUpTime":"5","rampUpSteps":"2"}`,
			wantMsgIn: "virtual user",
		},

		// --- in-range boundaries must pass range validation (stop at http request check) ---
		{
			name:      "in-range min boundary (1/1/1/1) passes range validation",
			body:      `{` + base + `,"virtualUsers":"1","duration":"1","rampUpTime":"1","rampUpSteps":"1","httpReqs":[]}`,
			wantMsgIn: "http request",
		},
		{
			name:      "in-range max boundary (100/300/60/20) passes range validation",
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
