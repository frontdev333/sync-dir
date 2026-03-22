package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

func main() {
	pth, err := getFilePath()
	if err != nil {
		slog.Error(err.Error())
		return
	}

	res, err := ListFiles(pth)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	for _, v := range res {
		fmt.Println(v)
	}
}

func getFilePath() (string, error) {
	if len(os.Args) < 2 {
		return "", errors.New("root path wasn't input")
	}

	pth := os.Args[1:]
	return pth[0], nil
}

func ListFiles(rootPath string) ([]string, error) {
	var res []string
	err := filepath.WalkDir(rootPath, func(pth string, d fs.DirEntry, err error) error {

		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		pth, err = filepath.Rel(rootPath, pth)
		if err != nil {
			slog.Error("file needs to be checked manually", "file", pth, "error", err)
			return err
		}
		res = append(res, pth)

		return nil
	})

	return res, err
}
