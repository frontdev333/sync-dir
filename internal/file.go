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
	"sync"
)

type FileResult struct {
	Path string
	Hash string
}

func GetFilePath() (string, error) {
	if len(os.Args) < 2 {
		return "", errors.New("root path wasn't input")
	}

	pth := os.Args[1:]
	return pth[0], nil
}

func ListFiles(res chan<- string, rootPath string) {
	defer close(res)
	err := filepath.WalkDir(rootPath, func(pth string, d fs.DirEntry, err error) error {

		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		res <- pth
		return nil
	})

	if err != nil {
		slog.Error(err.Error())
	}
}

func HashFile(wg *sync.WaitGroup, res chan<- FileResult, tasks <-chan string, rootPath string) {
	defer wg.Done()
	h := sha256.New()

	for pth := range tasks {
		h.Reset()

		file, err := os.Open(pth)
		if err != nil {
			slog.Error(err.Error())
			continue
		}

		if _, err = io.Copy(h, file); err != nil {
			file.Close()
			slog.Error(err.Error())
			continue
		}
		file.Close()

		hashInBytes := h.Sum(nil)
		hashRes := hex.EncodeToString(hashInBytes)

		pth, err = filepath.Rel(rootPath, pth)
		if err != nil {
			slog.Error("file needs to be checked manually", "file", pth, "error", err)
			continue
		}

		res <- FileResult{
			Path: pth,
			Hash: hashRes,
		}
	}
}
