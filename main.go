package main

import (
	"fmt"
	"frontdev333/sync-dir/internal"
	"log/slog"
)

func main() {
	pth, err := internal.GetFilePath()
	if err != nil {
		slog.Error(err.Error())
		return
	}

	res, err := internal.ListFiles(pth)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	for f, h := range res {
		fmt.Printf("%s: %s\n", f, h)
	}
}
