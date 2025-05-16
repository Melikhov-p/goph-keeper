package user_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Melikhov-p/goph-keeper/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	login    = "login"
	password = "password"
	pepper   = "supersecrethashforpepper"
)

func TestNewUser(t *testing.T) {
	testCases := []struct {
		name     string
		login    string
		password string
		wantErr  bool
	}{
		{
			name:     "success",
			login:    login,
			password: password,
			wantErr:  false,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			u, err := user.NewUser(test.login, test.password, pepper)

			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, test.login, u.Login)
			assert.Equal(t, time.Now(), u.CreatedAt)
		})
	}
}

func ExampleNewUser() {
	u, err := user.NewUser(login, password, pepper)
	if err != nil {
		panic("fail to create user")
	}
	fmt.Println(u.ID)

	// Output:
	// 0
}

func TestUser_VerifyUserPassword(t *testing.T) {
	u, err := user.NewUser(login, password, pepper)
	require.NoError(t, err)

	wrongPass := "wrong"

	testCases := []struct {
		name     string
		password string
		verified bool
	}{
		{
			name:     "success",
			password: password,
			verified: true,
		},
		{
			name:     "fail",
			password: wrongPass,
			verified: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			ok := u.VerifyUserPassword(test.password, pepper)

			if test.verified {
				require.True(t, ok)
			} else {
				require.False(t, ok)
			}
		})
	}
}
