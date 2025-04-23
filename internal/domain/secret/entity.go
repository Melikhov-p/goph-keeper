// Package secret пакет уровня домена секретов.
package secret

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"github.com/Melikhov-p/goph-keeper/internal/domain/user"
	"github.com/Melikhov-p/goph-keeper/internal/encryptor"
)

// TypeOfSecret тип секрета.
type TypeOfSecret string

type SecretData interface {
	Encrypt() error
	Decrypt() error
	setID(newID int)
}

var ErrInvalidSecretType = errors.New("invalid secret type")

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
// Изначально секрету присваивается ID = -1 после записи в БД ID меняется на присвоенный в базе.
type Secret struct {
	ID        int
	UserID    int
	Name      string
	Type      TypeOfSecret
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
	Version   uint32
	Data      SecretData
}

func NewSecret(secretName string, secretType TypeOfSecret, userID int) *Secret {
	now := time.Now()

	return &Secret{
		ID:        -1,
		UserID:    userID,
		Name:      secretName,
		Type:      secretType,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}
}

func (s *Secret) SetData(data SecretData) {
	s.Data = data
}

// Базовая структура данных секрета.
type baseSecretData struct {
	SecretID  int
	Notes     string
	MetaData  []byte
	Encrypted bool
}

func (s *Secret) SetID(newID int) {
	s.ID = newID
	s.Data.setID(newID)
}

// newBaseSecretData получение модели с данными базовыми для всех секретных данных.
// Изначально данные внутри секрета связываются с секретом ID = -1 после записи в БД ID меняется на присвоенный в базе.
func newBaseSecretData(secretID int, notes string, metaData []byte) *baseSecretData {
	return &baseSecretData{
		SecretID:  secretID,
		Notes:     notes,
		MetaData:  metaData,
		Encrypted: false,
	}
}

func (bs *baseSecretData) encrypt() error {
	op := "domain.service.baseSecretData.encrypt"

	var err error

	bs.Notes, err = encryptor.Encrypt([]byte(bs.Notes))
	if err != nil {
		return fmt.Errorf("%s: failed to encrypt secret notes %w", op, err)
	}

	return nil
}

func (bs *baseSecretData) decrypt() error {
	op := "domain.service.baseSecretData.encrypt"

	var err error

	bs.Notes, err = encryptor.Decrypt([]byte(bs.Notes))
	if err != nil {
		return fmt.Errorf("%s: failed to decrypt secret notes %w", op, err)
	}

	return nil
}

// PasswordData структура секрета для хранения пароля.
type PasswordData struct {
	*baseSecretData
	Username string
	Pass     string
	URL      string
}

// NewPasswordData получение новой модели для данных внутри секрета с паролем.
func NewPasswordData(
	secret *Secret,
	username, password, url, notes string,
	metaData []byte,
) *PasswordData {
	base := newBaseSecretData(secret.ID, notes, metaData)

	return &PasswordData{
		baseSecretData: base,
		Username:       username,
		Pass:           password,
		URL:            url,
	}
}

// NewPasswordSecret получение новой модели для секрета с паролем.
func NewPasswordSecret(
	secretName,
	username, password, url, notes string,
	u user.User,
	metaData []byte,
) (*Secret, error) {
	op := "domain.service.NewPasswordSecret"

	var (
		secret *Secret
		data   *PasswordData
		err    error
	)

	secret = NewSecret(secretName, TypePassword, u.ID)

	data = NewPasswordData(secret, username, password, url, notes, metaData)

	secret.SetData(data)

	err = secret.Data.Encrypt()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to encrypt PasswordData %w", op, err)
	}

	return secret, nil
}

func (pd *PasswordData) Encrypt() error {
	op := "domain.service.PasswordData.encrypt"

	var err error

	pd.Pass, err = encryptor.Encrypt([]byte(pd.Pass))
	if err != nil {
		return fmt.Errorf("%s: failed to encrypt password %w", op, err)
	}

	pd.Notes, err = encryptor.Encrypt([]byte(pd.Notes))
	if err != nil {
		return fmt.Errorf("%s: failed to encrypt password %w", op, err)
	}

	err = pd.baseSecretData.encrypt()
	if err != nil {
		return fmt.Errorf("%s: failed to encrypt base secret data %w", op, err)
	}

	pd.Encrypted = true

	return nil
}

func (pd *PasswordData) Decrypt() error {
	op := "domain.service.PasswordData.decrypt"

	var err error

	pd.Pass, err = encryptor.Decrypt([]byte(pd.Pass))
	if err != nil {
		return fmt.Errorf("%s: failed to encrypt password %w", op, err)
	}

	pd.Notes, err = encryptor.Decrypt([]byte(pd.Notes))
	if err != nil {
		return fmt.Errorf("%s: failed to encrypt password %w", op, err)
	}

	err = pd.baseSecretData.decrypt()
	if err != nil {
		return fmt.Errorf("%s: failed to decrypt base secret data %w", op, err)
	}

	pd.Encrypted = false

	return nil
}

func (pd *PasswordData) setID(newID int) {
	pd.SecretID = newID
}

// CardData структура для секрета с данными карты.
type CardData struct {
	*baseSecretData
	Number     string
	Owner      string
	ExpireDate string
	CVV        string
}

// NewCardData получение новой модели для данных внутри секрета с паролем.
func NewCardData(
	secret *Secret,
	number, owner, expireDate, cvv string,
	notes string,
	metaData []byte,
) *CardData {
	base := newBaseSecretData(secret.ID, notes, metaData)

	return &CardData{
		baseSecretData: base,
		Number:         number,
		Owner:          owner,
		CVV:            cvv,
	}
}

// NewCardSecret получение новой модели для секрета с паролем.
func NewCardSecret(
	secretName,
	number, owner, expireDate, cvv string,
	notes string,
	u user.User,
	metaData []byte,
) (*Secret, error) {
	op := "domain.service.NewPasswordSecret"

	var (
		secret *Secret
		data   *CardData
		err    error
	)

	secret = NewSecret(secretName, TypeCard, u.ID)

	data = NewCardData(secret, number, owner, expireDate, cvv, notes, metaData)

	secret.SetData(data)

	err = secret.Data.Encrypt()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to encrypt CardData %w", op, err)
	}

	return secret, nil
}

func (cd *CardData) Encrypt() error {
	op := "domain.service.PasswordData.encrypt"

	var err error

	cd.Number, err = encryptor.Encrypt([]byte(cd.Number))
	if err != nil {
		return fmt.Errorf("%s: failed to encrypt card number %w", op, err)
	}

	cd.Owner, err = encryptor.Encrypt([]byte(cd.Owner))
	if err != nil {
		return fmt.Errorf("%s: failed to encrypt card owner %w", op, err)
	}

	cd.CVV, err = encryptor.Encrypt([]byte(cd.CVV))
	if err != nil {
		return fmt.Errorf("%s: failed to encrypt card cvv %w", op, err)
	}

	err = cd.baseSecretData.encrypt()
	if err != nil {
		return fmt.Errorf("%s: failed to encrypt base secret data %w", op, err)
	}

	cd.Encrypted = true

	return nil
}

func (cd *CardData) Decrypt() error {
	op := "domain.service.PasswordData.decrypt"

	var err error

	cd.Number, err = encryptor.Decrypt([]byte(cd.Number))
	if err != nil {
		return fmt.Errorf("%s: failed to decrypt card number %w", op, err)
	}

	cd.Owner, err = encryptor.Decrypt([]byte(cd.Owner))
	if err != nil {
		return fmt.Errorf("%s: failed to decrypt card owner %w", op, err)
	}

	cd.CVV, err = encryptor.Decrypt([]byte(cd.CVV))
	if err != nil {
		return fmt.Errorf("%s: failed to decrypt card cvv %w", op, err)
	}

	err = cd.baseSecretData.decrypt()
	if err != nil {
		return fmt.Errorf("%s: failed to decrypt base secret data %w", op, err)
	}

	cd.Encrypted = false

	return nil
}

func (cd *CardData) setID(newID int) {
	cd.SecretID = newID
}

// FileData структура для секретных файлов.
type FileData struct {
	*baseSecretData
	Path    string
	Name    string
	Content string
}

// NewFileData получение новой модели для данных внутри секрета с паролем.
func NewFileData(
	secret *Secret,
	path, name, content string,
	notes string,
	metaData []byte,
) *FileData {
	base := newBaseSecretData(secret.ID, notes, metaData)

	return &FileData{
		baseSecretData: base,
		Path:           path,
		Name:           name,
		Content:        content,
	}
}

// NewFileSecret получение новой модели для секрета с паролем.
func NewFileSecret(
	secretName,
	path, name, content string,
	notes string,
	u user.User,
	metaData []byte,
) (*Secret, error) {
	op := "domain.service.NewPasswordSecret"

	var (
		secret *Secret
		data   *FileData
		err    error
	)

	secret = NewSecret(secretName, TypeCard, u.ID)

	data = NewFileData(secret, path, name, content, notes, metaData)

	secret.SetData(data)

	err = secret.Data.Encrypt()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to encrypt CardData %w", op, err)
	}

	return secret, nil
}

func (fd *FileData) Encrypt() error {
	op := "domain.service.PasswordData.encrypt"

	var err error

	fd.Content, err = encryptor.Encrypt([]byte(fd.Content))
	if err != nil {
		return fmt.Errorf("%s: failed to encrypt file content %w", op, err)
	}

	err = fd.baseSecretData.encrypt()
	if err != nil {
		return fmt.Errorf("%s: failed to encrypt base secret data %w", op, err)
	}

	fd.Encrypted = true

	return nil
}

func (fd *FileData) Decrypt() error {
	op := "domain.service.PasswordData.decrypt"

	var err error

	fd.Content, err = encryptor.Encrypt([]byte(fd.Content))
	if err != nil {
		return fmt.Errorf("%s: failed to decrypt file content %w", op, err)
	}

	err = fd.baseSecretData.decrypt()
	if err != nil {
		return fmt.Errorf("%s: failed to decrypt base secret data %w", op, err)
	}

	fd.Encrypted = false

	return nil
}

func (fd *FileData) setID(newID int) {
	fd.SecretID = newID
}
