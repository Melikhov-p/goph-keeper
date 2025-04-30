package external_storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// SaveFileData сохранить файл секретный, возвращает контрольную сумму файла, его полный путь и ошибку.
func SaveFileData(_ context.Context, userID int, path string, content []byte) (string, string, error) {
	op := "SaveFileData"

	var (
		hasher   hash.Hash
		out      *os.File
		checksum string
		err      error
	)

	_, err = os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", "", fmt.Errorf("%s: failed to check external storage path: it does not exist", op)
		}
		return "", "", fmt.Errorf("%s: failed to chekc external storage path with error %w", op, err)
	}

	// Создаем папку пользователя
	userDir := filepath.Join(path, "user_"+strconv.Itoa(userID))
	if err = os.MkdirAll(userDir, 0o700); err != nil {
		return "", "", fmt.Errorf("%s: failed to create user folder with error %w", op, err)
	}

	filePath := filepath.Join(userDir, time.Now().Format("2006_01_02_15_04_05"))

	// Сохраняем файл
	out, err = os.Create(filePath)
	if err != nil {
		return "", "", fmt.Errorf("%s: failed to create file with error %w", op, err)
	}
	defer func() {
		_ = out.Close()
	}()

	err = os.WriteFile(filePath, content, os.ModeAppend)
	if err != nil {
		return "", "", fmt.Errorf("%s: failed to write to file with error %w", op, err)
	}

	hasher = sha256.New()
	if _, err = io.Copy(hasher, out); err != nil {
		return "", "", fmt.Errorf("%s: failed to copy file content to hasher with error %w", op, err)
	}

	checksum = hex.EncodeToString(hasher.Sum(nil))

	return checksum, out.Name(), nil
}
