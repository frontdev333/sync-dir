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
	"sync/atomic"
)

const workersNum = 100

type FileResult struct {
	Path string
	Hash string
}

type SyncTask struct {
	Action string
	Src    string
	Dst    string
}

type FileCounter struct {
	Del atomic.Int32
	Upd atomic.Int32
	Cp  atomic.Int32
}

func GetPaths() (string, string, error) {
	if len(os.Args) < 3 {
		return "", "", errors.New("input source path and destination path")
	}

	pths := os.Args[1:]
	srcPth := pths[0]
	destPth := pths[1]
	return srcPth, destPth, nil
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

func CompareScans(source, dest map[string]string) (toCopy []string, toUpdate []string, toDelete []string) {
	for k, v := range source {
		h, ok := dest[k]
		if !ok {
			toCopy = append(toCopy, k)
			continue
		}
		if h != v {
			toUpdate = append(toUpdate, k)
			continue
		}
	}

	for k, _ := range dest {
		_, ok := source[k]
		if !ok {
			toDelete = append(toDelete, k)
			continue
		}
	}

	return toCopy, toUpdate, toDelete
}

func ScanDir(externalCh chan<- map[string]string, pth string) {
	defer close(externalCh)

	tasks := make(chan string)
	results := make(chan FileResult)
	finalMap := make(map[string]string)
	wg := &sync.WaitGroup{}

	for i := 0; i < workersNum; i++ {
		wg.Add(1)
		go HashFile(wg, results, tasks, pth)
	}

	go ListFiles(tasks, pth)
	go func() {
		wg.Wait()
		close(results)
	}()

	for v := range results {
		finalMap[v.Path] = v.Hash
	}
	externalCh <- finalMap
}

func CopyFile(srcPth, dstPth string) {
	srcFile, err := os.Open(srcPth)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer srcFile.Close()

	if err = os.MkdirAll(filepath.Dir(dstPth), 0755); err != nil {
		slog.Error(err.Error())
		return
	}

	dstFile, err := os.Create(dstPth)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		slog.Error(err.Error())
		return
	}
}

func DeleteFile(pth string) {
	if err := os.Remove(pth); err != nil {
		slog.Error(err.Error())
	}
}

func UpdateFile(srcPth, dstPth string) {
	DeleteFile(dstPth)
	CopyFile(srcPth, dstPth)
}
