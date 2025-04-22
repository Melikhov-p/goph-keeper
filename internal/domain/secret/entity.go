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

var InvalidSecretTypeErr = errors.New("invalid secret type")

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
	return nil
}

// Value реализует интерфейс driver.Valuer для записи в БД.
func (tos *TypeOfSecret) Value() (driver.Value, error) {
	return string(*tos), nil
}

// Secret структура секрета.
type Secret[T PasswordData | CardData | FileData] struct {
	ID        int
	UserID    int
	Name      string
	Type      TypeOfSecret
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
	Version   uint32
	Data      T
}

// NewSecret получение новой модели домена секрета.
func NewSecret[T PasswordData | CardData | FileData](userID int, name string, data any) (*Secret[T], error) {
	var secretData T

	secret := &Secret[T]{
		UserID:    userID,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	switch any(secretData).(type) {
	case PasswordData:
		secret.Type = TypePassword
	case CardData:
		secret.Type = TypeCard
	case FileData:
		secret.Type = TypeBinary
	default:
		return nil, InvalidSecretTypeErr
	}

	secret.Data = data

	return secret, nil
}

type baseSecretData struct {
	SecretID       int
	Username       string
	NotesEncrypted string
	MetaData       []byte
}

func newBaseSecretData(secretID int, username, notes string, metaData []byte) (baseSecretData, error) {
	op := "domain.baseSecretData.newBaseSecretData"

	var (
		notesEnc string
		err      error
	)

	notesEnc, err = encryptor.Encrypt([]byte(notes))
	if err != nil {
		return baseSecretData{}, fmt.Errorf("%s: %w", op, err)
	}

	return baseSecretData{
		SecretID:       secretID,
		Username:       username,
		NotesEncrypted: notesEnc,
		MetaData:       metaData,
	}, nil
}

// PasswordData структура секрета для хранения пароля.
type PasswordData struct {
	baseSecretData
	PassEncrypted  string
	URL            string
	NotesEncrypted string
	MetaData       []byte
}

// NewPasswordData получение новой модели секрета с паролем.
func NewPasswordData(secretID, userID int, username, pass, url, notes string, metaData []byte) (*Secret[PasswordData], error) {
	op := "domain.PasswordData.NewPasswordData"

	var (
		passEnc string
		secret  *Secret[PasswordData]
		data    *PasswordData
		err     error
	)

	base, err := newBaseSecretData(secretID, username, notes, metaData)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	passEnc, err = encryptor.Encrypt([]byte(pass))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	data = &PasswordData{
		baseSecretData: base,
		PassEncrypted:  passEnc,
		URL:            url,
	}

	secret, err = NewSecret[PasswordData](userID, username, data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return secret, nil
}

// CardData структура для секрета с данными карты.
type CardData struct {
	baseSecretData
	NumberEnc     string
	OwnerEnc      string
	ExpireDateEnc string
	CVVEnc        string
}

// NewCardData новая модель для секрета данных карты.
func NewCardData(
	secretID, userID int,
	username, number, owner, cvv, notes, expireDate string,
	metaData []byte,
) (*Secret[CardData], error) {
	op := "domain.CardData.NewCardData"

	var (
		data                                       *CardData
		secret                                     *Secret[CardData]
		numberEnc, ownerEnc, expireDateEnc, cvvEnc string
		err                                        error
	)

	base, err := newBaseSecretData(secretID, username, notes, metaData)
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

	data = &CardData{
		baseSecretData: base,
		NumberEnc:      numberEnc,
		OwnerEnc:       ownerEnc,
		ExpireDateEnc:  expireDateEnc,
		CVVEnc:         cvvEnc,
	}

	secret, err = NewSecret[CardData](userID, username, data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return secret, nil
}

// FileData структура для секретных файлов.
type FileData struct {
	baseSecretData
	Path       string
	Name       string
	ContentEnc string
}

// NewFileData новая модель секретного файла. (двоичных данных)
func NewFileData(
	secretID, userID int,
	username, fileName, notes, pathToSave string,
	content, metaData []byte,
) (*Secret[FileData], error) {
	op := "domain.FileData.NewFileData"

	var (
		data       *FileData
		secret     *Secret[FileData]
		contentEnc string
		err        error
	)

	base, err := newBaseSecretData(secretID, username, notes, metaData)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	contentEnc, err = encryptor.Encrypt(content)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	data = &FileData{
		baseSecretData: base,
		Path:           pathToSave + "/" + fileName,
		ContentEnc:     contentEnc,
		Name:           fileName,
	}

	secret, err = NewSecret[FileData](userID, username, data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return secret, nil
}
