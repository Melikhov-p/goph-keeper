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

const (
	masterKeyByteLen = 32
)

var (
	errInvalidKeyLength = errors.New("invalid key length: must be masterKeyByteLen bytes")
	errDecryptionFailed = errors.New("decryption failed")
)

// EncryptWithMasterKey шифрует данные с использованием мастер-ключа.
// Принимает: plaintext - данные для шифрования, masterKey - masterKeyByteLen-байтный ключ.
// Возвращает: строку в формате "encryptedKey:encryptedData" или ошибку.
func EncryptWithMasterKey(plaintext []byte, masterKey []byte) (string, error) {
	op := "encrypt.EncryptWithMasterKey"

	if len(masterKey) != masterKeyByteLen {
		return "", fmt.Errorf("%s: master key %w", op, errInvalidKeyLength)
	}

	// Генерируем случайный ключ для данных
	dataKey := make([]byte, masterKeyByteLen)
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

// DecryptWithMasterKey расшифровывает данные, используя мастер-ключ.
func DecryptWithMasterKey(encoded []byte, masterKey []byte) (string, error) {
	op := "encrypt.DecryptWithMasterKey"

	keyPartsCount := 2

	if len(masterKey) != masterKeyByteLen {
		return "", fmt.Errorf("%s: masterKey %w", op, errInvalidKeyLength)
	}

	parts := strings.Split(string(encoded), ":")
	if len(parts) != keyPartsCount {
		return "", errDecryptionFailed
	}

	// Расшифровываем ключ данных (получаем []byte)
	dataKeyBytes, err := decryptToBytes(parts[0], masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data key: %w", err)
	}

	// Проверяем длину ключа
	if len(dataKeyBytes) != masterKeyByteLen {
		return "", fmt.Errorf("%s: dataKeyBytes %w", op, errInvalidKeyLength)
	}

	// Расшифровываем данные
	plaintext, err := decrypt(parts[1], dataKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}

// encrypt выполняет AES-GCM шифрование.
func encrypt(plaintext []byte, key []byte) (string, error) {
	op := "encryptor.Encrypt.encrypt"

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("%s: failed to NewCipher with error %w", op, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("%s: failed to NewGCM with error %w", op, err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("%s: failed to read full with error %w", op, err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt выполняет AES-GCM дешифрование и возвращает строку.
func decrypt(encodedCiphertext string, key []byte) (string, error) {
	op := "encryptor.Encrypt.decrypt"

	plaintext, err := decryptToBytes(encodedCiphertext, key)
	if err != nil {
		return "", fmt.Errorf("%s: failed to decryptToBytes with error %w", op, err)
	}
	return string(plaintext), nil
}

// decryptToBytes выполняет AES-GCM дешифрование и возвращает []byte.
func decryptToBytes(encodedCiphertext string, key []byte) ([]byte, error) {
	op := "encryptor.Encrypt.decryptToBytes"

	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to DecodeString with error %w", op, err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to NewCipher with error %w", op, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to NewGCM with error %w", op, err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errDecryptionFailed
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
