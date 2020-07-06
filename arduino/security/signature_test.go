package security

import (
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestSignatureVerification(t *testing.T) {
	res, signer, err := VerifyArduinoDetachedSignature(paths.New("testdata/package_index.json"), paths.New("testdata/package_index.json.sig"))
	require.NoError(t, err)
	require.NotNil(t, signer)
	require.True(t, res)
	require.Equal(t, uint64(0x7baf404c2dfab4ae), signer.PrimaryKey.KeyId)
}
