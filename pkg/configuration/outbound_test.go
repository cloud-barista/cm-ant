package configuration

import (
	"testing"
)

func TestTumblebugHostWithPort(t *testing.T) {
	expected := "127.0.0.1:1323"
	result := TumblebugHostWithPort()
	if result != expected {
		t.Errorf("Expected: %s, Got: %s", expected, result)
	}
}

func TestTumblebugBaseAuthHeader(t *testing.T) {
	expected := "Basic ZGVmYXVsdDpkZWZhdWx0" // "default:default"를 base64 인코딩
	result := TumblebugBaseAuthHeader()
	if result != expected {
		t.Errorf("Expected: %s, Got: %s", expected, result)
	}
}
