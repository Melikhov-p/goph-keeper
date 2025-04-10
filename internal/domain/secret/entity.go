// Package secret пакет уровня домена секретов.
package secret

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"
)

// TypeOfSecret тип секрета.
type TypeOfSecret string

const (
	// TypePassword секрет пароль.
	TypePassword TypeOfSecret = "password"
	// TypeCard секрет карта.
	TypeCard TypeOfSecret = "card"
	// TypeNote секрет текст.
	TypeNote TypeOfSecret = "note"
	// TypeBinary секрет бинарный.
	TypeBinary TypeOfSecret = "binary"
)

// Valid проверка валидности типа секрета.
func (tos *TypeOfSecret) Valid() bool {
	switch *tos {
	case TypePassword, TypeCard, TypeNote, TypeBinary:
		return true
	default:
		return false
	}
}

// Scan реализует интерфейс sql.Scanner для чтения из БД.
func (tos *TypeOfSecret) Scan(value interface{}) error {
	if value == nil {
		return errors.New("secret type cannot be null")
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
		return nil, fmt.Errorf("cannot save invalid secret type: %s", string(*tos))
	}
	return string(*tos), nil
}

// Secret структура секрета.
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

// NewSecret получение новой модели домена секрета.
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

// PasswordData структура секрета для хранения пароля.
type PasswordData struct {
	SecretID       int
	Username       string
	PassEncrypted  string
	URL            string
	NotesEncrypted string
	MetaData       []byte
}
