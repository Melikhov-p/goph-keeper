// Package secret пакет уровня домена секретов.
package secret

import (
	"database/sql"
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
	setDataFromRow(row *sql.Row) error
}

var ErrInvalidSecretType = errors.New("invalid secret type")

const (
	// TypePassword секрет пароль.
	TypePassword TypeOfSecret = "password"
	// TypeCard секрет карта.
	TypeCard TypeOfSecret = "card"
	// TypeBinary секрет бинарный.
	TypeBinary TypeOfSecret = "binary"

	// TypeNote секрет текст.
	// TypeNote TypeOfSecret = "note"
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

// NewSecret получить новый секрет.
func NewSecret(secretName string, secretType TypeOfSecret, userID int) (*Secret, error) {
	var now time.Time

	now = time.Now()

	switch secretType {
	case TypePassword:
	case TypeBinary:
	case TypeCard:
	default:
		return nil, ErrInvalidSecretType
	}

	return &Secret{
		ID:        -1,
		UserID:    userID,
		Name:      secretName,
		Type:      secretType,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}, nil
}

func (s *Secret) setData(data SecretData) {
	s.Data = data
}

// SetDataFromRow установить секрету секретные данные из строки из БД.
func (s *Secret) SetDataFromRow(data *sql.Row) error {
	err := s.Data.setDataFromRow(data)
	if err != nil {
		return fmt.Errorf("error setting data for %s", s.Type)
	}

	return nil
}

// Базовая структура данных секрета.
type baseSecretData struct {
	Notes     string
	MetaData  []byte
	Encrypted bool
}

// newBaseSecretData получение модели с данными базовыми для всех секретных данных.
// Изначально данные внутри секрета связываются с секретом ID = -1 после записи в БД ID меняется на присвоенный в базе.
func newBaseSecretData(notes string, metaData []byte) *baseSecretData {
	return &baseSecretData{
		Notes:     notes,
		MetaData:  metaData,
		Encrypted: false,
	}
}

func (bs *baseSecretData) Encrypt() error {
	op := "domain.service.baseSecretData.encrypt"

	var err error

	bs.Notes, err = encryptor.Encrypt([]byte(bs.Notes))
	if err != nil {
		return fmt.Errorf("%s: failed to encrypt secret notes %w", op, err)
	}

	return nil
}

func (bs *baseSecretData) Decrypt() error {
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
	username, password, url, notes string,
	metaData []byte,
) *PasswordData {
	base := newBaseSecretData(notes, metaData)

	return &PasswordData{
		baseSecretData: base,
		Username:       username,
		Pass:           password,
		URL:            url,
	}
}

// NewPasswordSecret получение новой модели для секрета с паролем.
func NewPasswordSecret(
	u *user.User,
	secretName,
	username, password, url, notes string,
	metaData []byte,
) (*Secret, error) {
	op := "domain.service.NewPasswordSecret"

	var (
		secret *Secret
		data   *PasswordData
		err    error
	)

	secret, err = NewSecret(secretName, TypePassword, u.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get secret domain model %w", op, err)
	}

	data = NewPasswordData(username, password, url, notes, metaData)

	secret.setData(data)

	err = secret.Data.Encrypt()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to encrypt PasswordData %w", op, err)
	}

	return secret, nil
}

func (pd *PasswordData) setDataFromRow(row *sql.Row) error {
	if err := row.Scan(&pd.Username, &pd.Pass, &pd.URL, &pd.Notes, &pd.MetaData); err != nil {
		return fmt.Errorf("failed to scan row for password data with error %w", err)
	}
	pd.Encrypted = true
	return nil
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

	err = pd.baseSecretData.Encrypt()
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

	err = pd.baseSecretData.Decrypt()
	if err != nil {
		return fmt.Errorf("%s: failed to decrypt base secret data %w", op, err)
	}

	pd.Encrypted = false

	return nil
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
	number, owner, expireDate, cvv string,
	notes string,
	metaData []byte,
) *CardData {
	base := newBaseSecretData(notes, metaData)

	return &CardData{
		baseSecretData: base,
		Number:         number,
		Owner:          owner,
		ExpireDate:     expireDate,
		CVV:            cvv,
	}
}

// NewCardSecret получение новой модели для секрета с паролем.
func NewCardSecret(
	u user.User,
	secretName, number, owner, expireDate, cvv string,
	notes string,
	metaData []byte,
) (*Secret, error) {
	op := "domain.service.NewPasswordSecret"

	var (
		secret *Secret
		data   *CardData
		err    error
	)

	secret, err = NewSecret(secretName, TypeCard, u.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get secret domain model %w", op, err)
	}

	data = NewCardData(number, owner, expireDate, cvv, notes, metaData)

	secret.setData(data)

	err = secret.Data.Encrypt()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to encrypt CardData %w", op, err)
	}

	return secret, nil
}

func (cd *CardData) setDataFromRow(row *sql.Row) error {
	if err := row.Scan(&cd.Number, &cd.Owner, &cd.ExpireDate, &cd.CVV, &cd.Notes, &cd.MetaData); err != nil {
		return fmt.Errorf("failed to scan row for password data with error %w", err)
	}
	cd.Encrypted = true
	return nil
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

	err = cd.baseSecretData.Encrypt()
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

	err = cd.baseSecretData.Decrypt()
	if err != nil {
		return fmt.Errorf("%s: failed to decrypt base secret data %w", op, err)
	}

	cd.Encrypted = false

	return nil
}

// FileData структура для секретных файлов.
type FileData struct {
	*baseSecretData
	Path    string
	Name    string
	Content []byte
}

// NewFileData получение новой модели для данных внутри секрета с паролем.
func NewFileData(
	path, name string,
	content []byte,
	notes string,
	metaData []byte,
) *FileData {
	base := newBaseSecretData(notes, metaData)

	return &FileData{
		baseSecretData: base,
		Path:           path,
		Name:           name,
		Content:        content,
	}
}

// NewFileSecret получение новой модели для секрета с паролем.
func NewFileSecret(
	u user.User,
	secretName,
	path, name string,
	content []byte,
	notes string,
	metaData []byte,
) (*Secret, error) {
	op := "domain.service.NewFileSecret"

	var (
		secret *Secret
		data   *FileData
		err    error
	)

	secret, err = NewSecret(secretName, TypeBinary, u.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get secret domain model %w", op, err)
	}

	data = NewFileData(path, name, content, notes, metaData)

	secret.setData(data)

	err = secret.Data.Encrypt()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to encrypt CardData %w", op, err)
	}

	return secret, nil
}

func (fd *FileData) setDataFromRow(row *sql.Row) error {
	op := "domain.service.setDataFromRow"
	if err := row.Scan(&fd.Path); err != nil {
		return fmt.Errorf("failed to scan row for password data with error %w", err)
	}

	err := fd.getContentFromFile(fd.Path)
	if err != nil {
		return fmt.Errorf(
			"%s: failed to get content file from file with path %s and error %w",
			op, fd.Path, err)
	}

	fd.Encrypted = true
	return nil
}

func (fd *FileData) getContentFromFile(path string) error {
	return nil
}

func (fd *FileData) Encrypt() error {
	op := "domain.service.PasswordData.encrypt"

	var (
		contentEnc string
		err        error
	)

	contentEnc, err = encryptor.Encrypt(fd.Content)
	if err != nil {
		return fmt.Errorf("%s: failed to encrypt file content %w", op, err)
	}
	fd.Content = []byte(contentEnc)

	err = fd.baseSecretData.Encrypt()
	if err != nil {
		return fmt.Errorf("%s: failed to encrypt base secret data %w", op, err)
	}

	fd.Encrypted = true

	return nil
}

func (fd *FileData) Decrypt() error {
	op := "domain.service.PasswordData.decrypt"

	var (
		contentDec string
		err        error
	)

	contentDec, err = encryptor.Decrypt(fd.Content)
	if err != nil {
		return fmt.Errorf("%s: failed to decrypt file content %w", op, err)
	}
	fd.Content = []byte(contentDec)

	err = fd.baseSecretData.Decrypt()
	if err != nil {
		return fmt.Errorf("%s: failed to decrypt base secret data %w", op, err)
	}

	fd.Encrypted = false

	return nil
}
