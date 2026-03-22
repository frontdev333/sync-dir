package internal

import (
	"fmt"
	"log/slog"
)

func Run() {
	srcPth, destPth, err := GetPaths()
	if err != nil {
		slog.Error(err.Error())
		return
	}

	srcRes := make(chan map[string]string)
	destRes := make(chan map[string]string)

	go ScanDir(srcRes, srcPth)
	go ScanDir(destRes, destPth)

	srcMap := <-srcRes
	destMap := <-destRes

	toCopy, toUpdate, toDelete := CompareScans(srcMap, destMap)

	fmt.Println("файлы для копирования:")
	for _, v := range toCopy {
		fmt.Println("- " + v)
	}

	fmt.Println("файлы для обновления:")
	for _, v := range toUpdate {
		fmt.Println("- " + v)
	}

	fmt.Println("файлы для удаления:")
	for _, v := range toDelete {
		fmt.Println("- " + v)
	}
}
