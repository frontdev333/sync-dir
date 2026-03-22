package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

func GetFilePath() (string, error) {
	if len(os.Args) < 2 {
		return "", errors.New("root path wasn't input")
	}

	pth := os.Args[1:]
	return pth[0], nil
}

func ListFiles(rootPath string) (map[string]string, error) {
	res := make(map[string]string)
	err := filepath.WalkDir(rootPath, func(pth string, d fs.DirEntry, err error) error {

		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		fileHash, err := HashFile(pth)
		if err != nil {
			slog.Error("make file hash error")
			return err
		}

		pth, err = filepath.Rel(rootPath, pth)
		if err != nil {
			slog.Error("file needs to be checked manually", "file", pth, "error", err)
			return err
		}
		res[pth] = fileHash

		return nil
	})

	return res, err
}

func HashFile(pth string) (string, error) {
	h := sha256.New()

	file, err := os.Open(pth)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err = io.Copy(h, file); err != nil {
		return "", err
	}

	hashInBytes := h.Sum(nil)

	return hex.EncodeToString(hashInBytes), nil
}
