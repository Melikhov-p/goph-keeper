package user

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User основная модель для пользователя.
type User struct {
	ID        int
	Login     string
	PassHash  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUser создает нового пользователя.
func NewUser(login, password, pepper string) (*User, error) {
	var (
		hash []byte
		err  error
		op   = "domain.User.NewUser"
	)

	hash, err = bcrypt.GenerateFromPassword([]byte(password+pepper), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &User{
		Login:     login,
		PassHash:  string(hash),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// VerifyUserPassword верификация пароля пользователя
func (u *User) VerifyUserPassword(password, pepper string) bool {
	var err error

	err = bcrypt.CompareHashAndPassword([]byte(u.PassHash), []byte(password+pepper))
	return err == nil
}
