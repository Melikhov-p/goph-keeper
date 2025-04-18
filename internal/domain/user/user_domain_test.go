package user

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
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
			user, err := NewUser(test.login, test.password, pepper)

			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, test.login, user.Login)
			assert.Equal(t, time.Now(), user.CreatedAt)
		})
	}
}

func ExampleNewUser() {
	user, err := NewUser(login, password, pepper)
	if err != nil {
		panic("fail to create user")
	}
	fmt.Println(user.ID)

	// Output:
	// 1
}

func TestUser_VerifyUserPassword(t *testing.T) {
	user, err := NewUser(login, password, pepper)
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
			ok := user.VerifyUserPassword(test.password, pepper)

			if test.verified {
				require.True(t, ok)
			} else {
				require.False(t, ok)
			}
		})
	}
}
