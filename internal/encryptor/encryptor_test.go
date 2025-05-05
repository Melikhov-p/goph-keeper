package encryptor

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptor(t *testing.T) {
	md, err := hex.DecodeString("f8f2761b99775dac26e373e4942d6fd648f29325db7312158cc88205ff5e86b8")
	require.NoError(t, err)

	testCases := []struct {
		name         string
		plainText    []byte
		masterKeyEnc []byte
		masterKeyDec []byte
		wantEncErr   bool
		wantDecErr   bool
	}{
		{
			name:         "success",
			plainText:    []byte("hello world"),
			masterKeyEnc: md,
			masterKeyDec: md,
			wantEncErr:   false,
			wantDecErr:   false,
		},
		{
			name:         "invalid key length",
			plainText:    []byte("hello world"),
			masterKeyEnc: make([]byte, 33),
			masterKeyDec: make([]byte, 33),
			wantEncErr:   true,
			wantDecErr:   true,
		},
		{
			name:         "invalid key",
			plainText:    []byte("hello world"),
			masterKeyEnc: md,
			masterKeyDec: make([]byte, 32),
			wantEncErr:   false,
			wantDecErr:   true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			var (
				encoded string
				decoded string
			)
			encoded, err = EncryptWithMasterKey(test.plainText, test.masterKeyEnc)
			if test.wantEncErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			decoded, err = DecryptWithMasterKey([]byte(encoded), test.masterKeyDec)
			if test.wantDecErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if !test.wantEncErr && !test.wantDecErr {
				assert.Equal(t, string(test.plainText), decoded)
			}
		})
	}
}
