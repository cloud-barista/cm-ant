package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateSSHKeyPair(t *testing.T) {
	bits := 4096
	err := GenerateSSHKeyPair(bits, "./id_rsa", "./id_rsa.pub")
	if err != nil {
		t.Fatalf("Failed to generate SSH key pair: %v", err)
	}
	require.NoError(t, err)
}
