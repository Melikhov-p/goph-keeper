package auth_test

import (
	"testing"
	"time"

	"github.com/Melikhov-p/goph-keeper/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	secretKey = "a6176d686706fe9caf7c85281d7f4730"
	toketTTL  = 15 * time.Second
)

func TestBuildJWTToken(t *testing.T) {
	testCases := []struct {
		name    string
		userID  int
		wantErr bool
	}{
		{
			name:    "success",
			userID:  1,
			wantErr: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			var (
				err error
			)

			_, err = auth.BuildJWTToken(test.userID, secretKey, toketTTL)

			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetUserIDbyToken(t *testing.T) {
	validUserID := 1

	token, err := auth.BuildJWTToken(validUserID, secretKey, toketTTL)
	require.NoError(t, err)

	testCases := []struct {
		name    string
		token   string
		userID  int
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   token,
			userID:  validUserID,
			wantErr: false,
		},
		{
			name:    "invalid token",
			token:   "qporipqoejrqejbr",
			userID:  999,
			wantErr: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			var userID int
			userID, err = auth.GetUserIDbyToken(test.token, secretKey)

			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.userID, userID)
			}
		})
	}
}
