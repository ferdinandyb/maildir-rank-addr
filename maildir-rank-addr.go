package main

import (
	"fmt"
	"github.com/emersion/go-message/mail"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

func processHeaders(headers chan *mail.Header) {
	count := 0
	for h := range headers {
		count++
		fmt.Println(count)
		fmt.Println(h.Subject())
	}
}

func walkMaildir(path string) {
	headers := make(chan *mail.Header)

	concurrent := runtime.GOMAXPROCS(2 * runtime.NumCPU())
	semaphore := make(chan struct{}, concurrent)
	var wg sync.WaitGroup
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		if info.IsDir() {
			return nil
		}
		switch filepath.Base(filepath.Dir(path)) {
		case "new", "tmp", "cur":
			wg.Add(1)
			go parseMessage(path, headers, semaphore, &wg)
		default:
			return nil
		}
		return nil
	})
	go processHeaders(headers)
	wg.Wait()

}

func main() {
	path := "/home/fbence/.mail"

	walkMaildir(path)
	// parseMessage(path)
}
