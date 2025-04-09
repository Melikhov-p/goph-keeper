package secret

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type TypeOfSecret string

const (
	TypePassword TypeOfSecret = "password"
	TypeCard     TypeOfSecret = "card"
	TypeNote     TypeOfSecret = "note"
	TypeBinary   TypeOfSecret = "binary"
	TypeOTP      TypeOfSecret = "otp"
)

func (tos *TypeOfSecret) Valid() bool {
	switch *tos {
	case TypePassword, TypeCard, TypeNote, TypeBinary, TypeOTP:
		return true
	default:
		return false
	}
}

// Scan реализует интерфейс sql.Scanner для чтения из БД.
func (tos *TypeOfSecret) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("secret type cannot be null")
	}

	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string for SecretType, got: %T", value)
	}

	*tos = TypeOfSecret(s)
	if !tos.Valid() {
		return fmt.Errorf("invalid secret type: %s", s)
	}
	return nil
}

// Value реализует интерфейс driver.Valuer для записи в БД.
func (tos *TypeOfSecret) Value() (driver.Value, error) {
	if !tos.Valid() {
		return nil, fmt.Errorf("cannot save invalid secret type: %s", tos)
	}
	return string(*tos), nil
}

type Secret struct {
	ID        int
	UserID    int
	Name      string
	Type      TypeOfSecret
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
	Version   uint32
}

func NewSecret(userID int, name string, t TypeOfSecret) (*Secret, error) {
	op := "domain.Secret.NewSecret"

	if !t.Valid() {
		return nil, fmt.Errorf("%s: invalid type of secret", op)
	}
	return &Secret{
		UserID:    userID,
		Name:      name,
		Type:      t,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}, nil
}

type PasswordData struct {
	SecretID       int
	Username       string
	PassEncrypted  string
	URL            string
	NotesEncrypted string
	MetaData       []byte
}

func NewPasswordSecret(
	userID int,
	name, username, password, pepper, url, notes string,
	meta []byte,
) (*Secret, *PasswordData, error) {
	op := "domain.Secret.NewPasswordSecret"

	var (
		secret    *Secret
		passData  *PasswordData
		passHash  []byte
		notesHash []byte
		err       error
	)

	if username == "" || password == "" {
		return nil, nil, fmt.Errorf("%s: username and password are required", op)
	}

	secret, err = NewSecret(userID, name, TypePassword)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: failed to get new domain model of secret %w", op, err)
	}

	passData = &PasswordData{
		SecretID:      secret.ID,
		Username:      username,
		PassEncrypted: password, // TODO: зашифровать пароль
		URL:           url,
	}
}
