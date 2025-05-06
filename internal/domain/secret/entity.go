// Package secret пакет уровня домена секретов.
package secret

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Melikhov-p/goph-keeper/internal/domain/user"
	"github.com/Melikhov-p/goph-keeper/internal/encryptor"
)

// TypeOfSecret тип секрета.
type TypeOfSecret string

// SecretData интерфейс секретных данных внутри секрета.
type SecretData interface {
	Encrypt() error
	Decrypt() error
	setDataFromRow(row *sql.Row) error
	setMasterKey(mk []byte)
}

var errInvalidSecretType = errors.New("invalid secret type")

const (
	// TypePassword секрет пароль.
	TypePassword TypeOfSecret = "password"
	// TypeCard секрет карта.
	TypeCard TypeOfSecret = "card"
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

// NewSecret получить новый секрет.
func NewSecret(secretName string, secretType TypeOfSecret, userID int) (*Secret, error) {
	now := time.Now()

	switch secretType {
	case TypePassword:
	case TypeBinary:
	case TypeCard:
	default:
		return nil, errInvalidSecretType
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
	switch s.Type {
	case TypePassword:
		s.Data = newEmptyPasswordData()
	case TypeCard:
		s.Data = newEmptyCardData()
	case TypeBinary:
		s.Data = newEmptyFileData()
	default:
		return fmt.Errorf("invalid secret type %s", s.Type)
	}

	if data == nil {
		return errors.New("data row is empty")
	}

	err := s.Data.setDataFromRow(data)
	if err != nil {
		return fmt.Errorf("error setting data for %s", s.Type)
	}

	return nil
}

// DecryptData расшифровать данные.
func (s *Secret) DecryptData() error {
	err := s.Data.Decrypt()
	if err != nil {
		return fmt.Errorf("failed to decrypt secret data with error %w", err)
	}

	return nil
}

// Базовая структура данных секрета.
type baseSecretData struct {
	Notes     string
	MetaData  []byte
	Encrypted bool
	masterKey []byte
}

// newBaseSecretData получение модели с данными базовыми для всех секретных данных.
// Изначально данные внутри секрета связываются с секретом ID = -1 после записи в БД ID меняется на присвоенный в базе.
func newBaseSecretData(notes string, metaData []byte, mk []byte) *baseSecretData {
	return &baseSecretData{
		Notes:     notes,
		MetaData:  metaData,
		Encrypted: false,
		masterKey: mk,
	}
}

func newEmptyBaseSecretData() *baseSecretData {
	return &baseSecretData{
		Notes:     "",
		MetaData:  nil,
		Encrypted: true,
		masterKey: nil,
	}
}

func (bs *baseSecretData) setMasterKey(mk []byte) {
	bs.masterKey = mk
}

func (bs *baseSecretData) Encrypt() error {
	op := "domain.service.baseSecretData.encrypt"

	var err error

	bs.Notes, err = encryptor.EncryptWithMasterKey([]byte(bs.Notes), bs.masterKey)
	if err != nil {
		return fmt.Errorf("%s: failed to encrypt secret notes %w", op, err)
	}

	return nil
}

func (bs *baseSecretData) Decrypt() error {
	op := "domain.service.baseSecretData.encrypt"

	if bs.Notes == "" {
		return nil
	}

	var err error

	bs.Notes, err = encryptor.DecryptWithMasterKey([]byte(bs.Notes), bs.masterKey)
	if err != nil {
		return fmt.Errorf("%s: failed to decrypt secret notes %w", op, err)
	}

	return nil
}

// PasswordData структура секрета для хранения пароля.
type PasswordData struct {
	*baseSecretData
	Username  string
	Pass      string
	URL       string
	masterKey []byte
}

// NewPasswordData получение новой модели для данных внутри секрета с паролем.
func NewPasswordData(
	username, password, url, notes string,
	metaData []byte,
	masterKey []byte,
) *PasswordData {
	base := newBaseSecretData(notes, metaData, masterKey)

	return &PasswordData{
		baseSecretData: base,
		Username:       username,
		Pass:           password,
		URL:            url,
		masterKey:      masterKey,
	}
}

// NewPasswordSecret получение новой модели для секрета с паролем.
func NewPasswordSecret(
	u *user.User,
	secretName,
	username, password, url, notes string,
	metaData []byte,
	masterKey []byte,
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

	data = NewPasswordData(username, password, url, notes, metaData, masterKey)

	secret.setData(data)

	err = secret.Data.Encrypt()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to encrypt PasswordData %w", op, err)
	}

	return secret, nil
}

func newEmptyPasswordData() *PasswordData {
	return &PasswordData{
		baseSecretData: newEmptyBaseSecretData(),
		Username:       "",
		Pass:           "",
		URL:            "",
	}
}

func (pd *PasswordData) setMasterKey(mk []byte) {
	pd.masterKey = mk
	pd.baseSecretData.setMasterKey(mk)
}

func (pd *PasswordData) setDataFromRow(row *sql.Row) error {
	if err := row.Scan(&pd.Username, &pd.Pass, &pd.URL, &pd.Notes, &pd.MetaData); err != nil {
		return fmt.Errorf("failed to scan row for password data with error %w", err)
	}
	pd.Encrypted = true
	return nil
}

// Encrypt шифрование данных секрета.
func (pd *PasswordData) Encrypt() error {
	op := "domain.service.PasswordData.encrypt"

	var err error

	pd.Pass, err = encryptor.EncryptWithMasterKey([]byte(pd.Pass), pd.masterKey)
	if err != nil {
		return fmt.Errorf("%s: failed to encrypt password %w", op, err)
	}

	pd.Notes, err = encryptor.EncryptWithMasterKey([]byte(pd.Notes), pd.masterKey)
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

// Decrypt расшифровать данные.
func (pd *PasswordData) Decrypt() error {
	op := "domain.service.PasswordData.decrypt"

	var err error

	pd.Pass, err = encryptor.DecryptWithMasterKey([]byte(pd.Pass), pd.masterKey)
	if err != nil {
		return fmt.Errorf("%s: failed to decrypt password %w", op, err)
	}

	pd.Notes, err = encryptor.DecryptWithMasterKey([]byte(pd.Notes), pd.masterKey)
	if err != nil {
		return fmt.Errorf("%s: failed to decrypt notes %w", op, err)
	}

	err = pd.baseSecretData.Decrypt()
	if err != nil {
		return fmt.Errorf("%s: error decoding base secret data %w", op, err)
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
	masterKey  []byte
}

// NewCardData получение новой модели для данных внутри секрета с паролем.
func NewCardData(
	number, owner, expireDate, cvv string,
	notes string,
	metaData []byte,
	masterKey []byte,
) *CardData {
	base := newBaseSecretData(notes, metaData, masterKey)

	return &CardData{
		baseSecretData: base,
		Number:         number,
		Owner:          owner,
		ExpireDate:     expireDate,
		CVV:            cvv,
		masterKey:      masterKey,
	}
}

func newEmptyCardData() *CardData {
	return &CardData{
		baseSecretData: newEmptyBaseSecretData(),
		Number:         "",
		Owner:          "",
		ExpireDate:     "",
		CVV:            "",
	}
}

func (cd *CardData) setMasterKey(mk []byte) {
	cd.masterKey = mk
	cd.baseSecretData.setMasterKey(mk)
}

// NewCardSecret получение новой модели для секрета с паролем.
func NewCardSecret(
	u *user.User,
	secretName, number, owner, expireDate, cvv string,
	notes string,
	metaData []byte,
	masterKey []byte,
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

	data = NewCardData(number, owner, expireDate, cvv, notes, metaData, masterKey)

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

// Encrypt шифрование данных.
func (cd *CardData) Encrypt() error {
	op := "domain.service.CardData.encrypt"

	var err error

	cd.Number, err = encryptor.EncryptWithMasterKey([]byte(cd.Number), cd.masterKey)
	if err != nil {
		return fmt.Errorf("%s: failed to encrypt card number %w", op, err)
	}

	cd.Owner, err = encryptor.EncryptWithMasterKey([]byte(cd.Owner), cd.masterKey)
	if err != nil {
		return fmt.Errorf("%s: failed to encrypt card owner %w", op, err)
	}

	cd.CVV, err = encryptor.EncryptWithMasterKey([]byte(cd.CVV), cd.masterKey)
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

// Decrypt дешифровка данных.
func (cd *CardData) Decrypt() error {
	op := "domain.service.PasswordData.decrypt"

	var err error

	cd.Number, err = encryptor.DecryptWithMasterKey([]byte(cd.Number), cd.masterKey)
	if err != nil {
		return fmt.Errorf("%s: failed to decrypt card number %w", op, err)
	}

	cd.Owner, err = encryptor.DecryptWithMasterKey([]byte(cd.Owner), cd.masterKey)
	if err != nil {
		return fmt.Errorf("%s: failed to decrypt card owner %w", op, err)
	}

	cd.CVV, err = encryptor.DecryptWithMasterKey([]byte(cd.CVV), cd.masterKey)
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
	Path      string
	Name      string
	Content   []byte
	masterKey []byte
}

// NewFileData получение новой модели для данных внутри секрета с паролем.
func NewFileData(
	path, name string,
	content []byte,
	notes string,
	metaData []byte,
	masterKey []byte,
) *FileData {
	base := newBaseSecretData(notes, metaData, masterKey)

	return &FileData{
		baseSecretData: base,
		Path:           path,
		Name:           name,
		Content:        content,
		masterKey:      masterKey,
	}
}

func newEmptyFileData() *FileData {
	return &FileData{
		baseSecretData: newEmptyBaseSecretData(),
		Path:           "",
		Name:           "",
		Content:        nil,
	}
}

func (fd *FileData) setMasterKey(mk []byte) {
	fd.masterKey = mk
	fd.baseSecretData.setMasterKey(mk)
}

// NewFileSecret получение новой модели для секрета с паролем.
func NewFileSecret(
	u *user.User,
	secretName,
	path, name string,
	content []byte,
	notes string,
	metaData []byte,
	masterKey []byte,
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

	data = NewFileData(path, name, content, notes, metaData, masterKey)

	secret.setData(data)

	err = secret.Data.Encrypt()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to encrypt CardData %w", op, err)
	}

	return secret, nil
}

func (fd *FileData) setDataFromRow(row *sql.Row) error {
	op := "domain.service.setDataFromRow"

	var err error

	if err = row.Scan(&fd.Path, &fd.Name); err != nil {
		return fmt.Errorf("failed to scan row for password data with error %w", err)
	}

	fd.Content, err = fd.GetContentFromFile()
	if err != nil {
		return fmt.Errorf(
			"%s: failed to get content file from file with path %s and error %w",
			op, fd.Path, err)
	}

	fd.Encrypted = true
	return nil
}

// GetContentFromFile получение содержимого секретного файла.
func (fd *FileData) GetContentFromFile() ([]byte, error) {
	content, err := os.ReadFile(fd.Path)
	if err != nil {
		return nil, fmt.Errorf("error reading content from file %w", err)
	}

	return content, nil
}

// Encrypt шифрование данных.
func (fd *FileData) Encrypt() error {
	op := "domain.service.FileData.encrypt"

	var (
		contentEnc string
		err        error
	)

	contentEnc, err = encryptor.EncryptWithMasterKey(fd.Content, fd.masterKey)
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

// Decrypt дешифровка данных.
func (fd *FileData) Decrypt() error {
	op := "domain.service.FileData.decrypt"

	var (
		contentDec string
		err        error
	)

	contentDec, err = encryptor.DecryptWithMasterKey(fd.Content, fd.masterKey)
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
