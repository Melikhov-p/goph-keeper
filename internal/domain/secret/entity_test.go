package secret_test

import (
	"encoding/hex"
	"testing"

	"github.com/Melikhov-p/goph-keeper/internal/domain/secret"
	"github.com/Melikhov-p/goph-keeper/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSecret(t *testing.T) {
	testCases := []struct {
		name       string
		secretName string
		secretType secret.TypeOfSecret
		userID     int
		wantErr    bool
	}{
		{
			name:       "success password",
			secretName: "password",
			secretType: secret.TypeOfSecret("password"),
			userID:     1,
			wantErr:    false,
		},
		{
			name:       "success card",
			secretName: "card",
			secretType: secret.TypeOfSecret("card"),
			userID:     1,
			wantErr:    false,
		},
		{
			name:       "success binary",
			secretName: "binary",
			secretType: secret.TypeOfSecret("binary"),
			userID:     1,
			wantErr:    false,
		},
		{
			name:       "invalid secret type",
			secretName: "invalid",
			secretType: secret.TypeOfSecret("invalid"),
			userID:     1,
			wantErr:    true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			s, err := secret.NewSecret(test.secretName, test.secretType, test.userID)

			if !test.wantErr {
				require.NoError(t, err)
				assert.Equal(t, test.secretType, s.Type)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestNewPasswordSecret(t *testing.T) {
	u, err := user.NewUser("test", "test", "test")
	require.NoError(t, err)

	mk, err := hex.DecodeString("f8f2761b99775dac26e373e4942d6fd648f29325db7312158cc88205ff5e86b8")
	require.NoError(t, err)

	testCases := []struct {
		name       string
		secretName string
		user       *user.User
		username   string
		password   string
		url        string
		notes      string
		wantErr    bool
	}{
		{
			name:       "success",
			secretName: "pass",
			user:       u,
			username:   "test",
			password:   "test",
			url:        "test",
			notes:      "test",
			wantErr:    false,
		},
		{
			name:       "fail",
			secretName: "pass",
			user:       u,
			username:   "test",
			password:   "",
			url:        "test",
			notes:      "test",
			wantErr:    true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			var s *secret.Secret
			s, err = secret.NewPasswordSecret(
				test.user,
				test.name,
				test.username,
				test.password,
				test.url,
				test.notes,
				make([]byte, 0),
				mk)

			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, secret.TypePassword, s.Type)
			}
		})
	}
}

func TestNewCardSecret(t *testing.T) {
	u, err := user.NewUser("test", "test", "test")
	require.NoError(t, err)

	mk, err := hex.DecodeString("f8f2761b99775dac26e373e4942d6fd648f29325db7312158cc88205ff5e86b8")
	require.NoError(t, err)

	testCases := []struct {
		name       string
		secretName string
		user       *user.User
		number     string
		owner      string
		expireDate string
		CVV        string
		notes      string
		wantErr    bool
	}{
		{
			name:       "success",
			secretName: "card",
			user:       u,
			number:     "123456758",
			owner:      "iam",
			expireDate: "01.23",
			CVV:        "123",
			notes:      "",
			wantErr:    false,
		},
		{
			name:       "fail",
			secretName: "card",
			user:       u,
			number:     "",
			owner:      "iam",
			expireDate: "01.23",
			CVV:        "123",
			notes:      "",
			wantErr:    true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			var cs *secret.Secret
			cs, err = secret.NewCardSecret(
				u,
				test.secretName,
				test.number,
				test.owner,
				test.expireDate,
				test.CVV,
				test.notes,
				make([]byte, 0),
				mk,
			)

			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, secret.TypeCard, cs.Type)
			}
		})
	}
}

func TestNewFileSecret(t *testing.T) {
	u, err := user.NewUser("test", "test", "test")
	require.NoError(t, err)

	mk, err := hex.DecodeString("f8f2761b99775dac26e373e4942d6fd648f29325db7312158cc88205ff5e86b8")
	require.NoError(t, err)

	testCases := []struct {
		name       string
		secretName string
		user       *user.User
		path       string
		fileName   string
		content    []byte
		notes      string
		wantErr    bool
	}{
		{
			name:       "success",
			secretName: "card",
			user:       u,
			path:       "path/to/file",
			fileName:   "iam",
			content:    []byte("hello world"),
			notes:      "",
			wantErr:    false,
		},
		{
			name:       "fail",
			secretName: "card",
			user:       u,
			path:       "path/to/file",
			fileName:   "iam",
			content:    make([]byte, 0),
			notes:      "",
			wantErr:    true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			var cs *secret.Secret
			cs, err = secret.NewFileSecret(
				u,
				test.secretName,
				test.path,
				test.fileName,
				test.content,
				test.notes,
				make([]byte, 0),
				mk,
			)

			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, secret.TypeBinary, cs.Type)
			}
		})
	}
}
