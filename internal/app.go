package internal

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"

	"github.com/schollz/progressbar/v3"
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

	tasks := make(chan SyncTask)
	wg := &sync.WaitGroup{}
	counter := &FileCounter{}
	bar := progressbar.Default(int64(len(toCopy) + len(toDelete) + len(toUpdate)))

	for i := 0; i < workersNum; i++ {
		wg.Add(1)
		go Worker(wg, tasks, counter, bar)
	}

	fmt.Println("Синхронизация...")
	for _, v := range toCopy {
		copyDest := filepath.Join(destPth, v)
		v = filepath.Join(srcPth, v)

		tasks <- SyncTask{
			Action: "Copy",
			Src:    v,
			Dst:    copyDest,
		}
	}

	for _, v := range toUpdate {
		updateSrc := filepath.Join(srcPth, v)
		v = filepath.Join(destPth, v)
		tasks <- SyncTask{
			Action: "Update",
			Src:    updateSrc,
			Dst:    v,
		}
	}

	for _, v := range toDelete {
		v = filepath.Join(destPth, v)
		tasks <- SyncTask{
			Action: "Delete",
			Src:    "",
			Dst:    v,
		}
	}
	close(tasks)

	wg.Wait()
	fmt.Println("Синхронизация завершена!")
	fmt.Println("Скопировано:", counter.Cp.Load())
	fmt.Println("Обновлено:", counter.Upd.Load())
	fmt.Println("Удалено:", counter.Del.Load())
}
