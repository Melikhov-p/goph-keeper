// Package encryptor пакет шифровальщика для секретов.
package encryptor

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	ErrInvalidKeyLength = errors.New("invalid key length: must be 32 bytes")
	ErrDecryptionFailed = errors.New("decryption failed")
)

// GenerateKey генерирует случайный 32-байтный ключ для AES-256
func GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

// EncryptWithMasterKey шифрует данные с использованием мастер-ключа.
// Принимает: plaintext - данные для шифрования, masterKey - 32-байтный ключ.
// Возвращает: строку в формате "encryptedKey:encryptedData" или ошибку.
func EncryptWithMasterKey(plaintext []byte, masterKey []byte) (string, error) {
	op := "encrypt.EncryptWithMasterKey"

	if len(masterKey) != 32 {
		return "", fmt.Errorf("%s: master key %w", op, ErrInvalidKeyLength)
	}

	// Генерируем случайный ключ для данных
	dataKey := make([]byte, 32)
	if _, err := rand.Read(dataKey); err != nil {
		return "", fmt.Errorf("failed to generate data key: %w", err)
	}

	// Шифруем данные
	encryptedData, err := encrypt(plaintext, dataKey)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt data: %w", err)
	}

	// Шифруем ключ данных
	encryptedKey, err := encrypt(dataKey, masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt data key: %w", err)
	}

	return encryptedKey + ":" + encryptedData, nil
}

// DecryptWithMasterKey расшифровывает данные, используя мастер-ключ
func DecryptWithMasterKey(encoded []byte, masterKey []byte) (string, error) {
	op := "encrypt.DecryptWithMasterKey"

	if len(masterKey) != 32 {
		return "", fmt.Errorf("%s: masterKey %w", op, ErrInvalidKeyLength)
	}

	parts := strings.Split(string(encoded), ":")
	if len(parts) != 2 {
		return "", ErrDecryptionFailed
	}

	// Расшифровываем ключ данных (получаем []byte)
	dataKeyBytes, err := decryptToBytes(parts[0], masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data key: %w", err)
	}

	// Проверяем длину ключа
	if len(dataKeyBytes) != 32 {
		return "", fmt.Errorf("%s: dataKeyBytes %w", op, ErrInvalidKeyLength)
	}

	// Расшифровываем данные
	plaintext, err := decrypt(parts[1], dataKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}

// encrypt выполняет AES-GCM шифрование
func encrypt(plaintext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt выполняет AES-GCM дешифрование и возвращает строку
func decrypt(encodedCiphertext string, key []byte) (string, error) {
	plaintext, err := decryptToBytes(encodedCiphertext, key)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// decryptToBytes выполняет AES-GCM дешифрование и возвращает []byte
func decryptToBytes(encodedCiphertext string, key []byte) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrDecryptionFailed
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
