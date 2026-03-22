package internal

import (
	"fmt"
	"log/slog"
	"sync"
)

const workersNum = 100

func Run() {
	tasks := make(chan string)
	results := make(chan FileResult)
	wg := &sync.WaitGroup{}

	pth, err := GetFilePath()
	if err != nil {
		slog.Error(err.Error())
		return
	}

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
		fmt.Printf("%s: %s\n", v.Path, v.Hash)
	}

}
