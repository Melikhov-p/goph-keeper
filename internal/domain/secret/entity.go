// Package secret пакет уровня домена секретов.
package secret

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"github.com/Melikhov-p/goph-keeper/internal/encryptor"
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

type BaseSecretData struct {
	SecretID       int
	Username       string
	NotesEncrypted string
	MetaData       []byte
}

func NewBaseSecretData(secretID int, username, notes string, metaData []byte) (BaseSecretData, error) {
	op := "domain.BaseSecretData.NewBaseSecretData"

	var (
		notesEnc string
		err      error
	)

	notesEnc, err = encryptor.Encrypt([]byte(notes))
	if err != nil {
		return BaseSecretData{}, fmt.Errorf("%s: %w", op, err)
	}

	return BaseSecretData{
		SecretID:       secretID,
		Username:       username,
		NotesEncrypted: notesEnc,
		MetaData:       metaData,
	}, nil
}

// PasswordData структура секрета для хранения пароля.
type PasswordData struct {
	BaseSecretData
	PassEncrypted  string
	URL            string
	NotesEncrypted string
	MetaData       []byte
}

// NewPasswordData получение новой модели секрета с паролем.
func NewPasswordData(secretID int, username, pass, url, notes string, metaData []byte) (*PasswordData, error) {
	op := "domain.PasswordData.NewPasswordData"

	var (
		passEnc string
		err     error
	)

	base, err := NewBaseSecretData(secretID, username, notes, metaData)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	passEnc, err = encryptor.Encrypt([]byte(pass))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &PasswordData{
		BaseSecretData: base,
		PassEncrypted:  passEnc,
		URL:            url,
	}, nil
}

// CardData структура для секрета с данными карты.
type CardData struct {
	BaseSecretData
	NumberEnc     string
	OwnerEnc      string
	ExpireDateEnc string
	CVVEnc        string
}

// NewCardData новая модель для секрета данных карты.
func NewCardData(
	secretID int,
	username, number, owner, cvv, notes, expireDate string,
	metaData []byte,
) (*CardData, error) {
	op := "domain.CardData.NewCardData"

	var (
		numberEnc, ownerEnc, expireDateEnc, cvvEnc string
		err                                        error
	)

	base, err := NewBaseSecretData(secretID, username, notes, metaData)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	numberEnc, err = encryptor.Encrypt([]byte(number))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	ownerEnc, err = encryptor.Encrypt([]byte(owner))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	expireDateEnc, err = encryptor.Encrypt([]byte(expireDate))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	cvvEnc, err = encryptor.Encrypt([]byte(cvv))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &CardData{
		BaseSecretData: base,
		NumberEnc:      numberEnc,
		OwnerEnc:       ownerEnc,
		ExpireDateEnc:  expireDateEnc,
		CVVEnc:         cvvEnc,
	}, nil
}

// FileData структура для секретных файлов.
type FileData struct {
	BaseSecretData
	Path       string
	Name       string
	ContentEnc string
}

// NewFileData новая модель секретного файла. (двоичных данных)
func NewFileData(secretID int, username, fileName, notes, pathToSave string, content, metaData []byte) (*FileData, error) {
	op := "domain.FileData.NewFileData"

	var (
		contentEnc string
		err        error
	)

	base, err := NewBaseSecretData(secretID, username, notes, metaData)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	contentEnc, err = encryptor.Encrypt(content)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &FileData{
		BaseSecretData: base,
		Path:           pathToSave + "/" + fileName,
		ContentEnc:     contentEnc,
		Name:           fileName,
	}, nil
}
