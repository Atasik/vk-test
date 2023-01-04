package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

func main() {
	// вынес все важные переменные в main, чтобы легче было менять код
	var total uint64 = 0
	textToFind := "Go"
	paths := make(chan string)
	wg := &sync.WaitGroup{}

	go CountData(paths, textToFind, wg, &total)
	ReadData(paths)

	wg.Wait()
	close(paths)
	fmt.Printf("Total: %d", total)
}

func ReadData(paths chan string) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		path := scanner.Text()
		switch path {
		case "":
			continue
		default:
			paths <- path
		}
	}
}

func CountData(paths chan string, textToFind string, wg *sync.WaitGroup, total *uint64) {
	goroutinesNum := 5
	waitCh := make(chan struct{}, goroutinesNum) // канал, который блокирует добавление "лишних" горутин

	for p := range paths {
		waitCh <- struct{}{}

		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			qty := сountSource(p, textToFind)
			atomic.AddUint64(total, qty) // использую atomic, чтобы писать в total из горутины
			fmt.Printf("Count for %s: %d\n", p, qty)
			<-waitCh
		}(p)
	}
}

func isUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func сountSource(path, substr string) uint64 {
	var reader io.Reader

	// проверяю, url или путь до файла
	if isUrl(path) {
		resp, err := http.Get(path)
		if err != nil {
			fmt.Println(err)
			return 0
		}

		defer resp.Body.Close()
		reader = bufio.NewReader(resp.Body)
	} else {
		file, err := os.Open(path)
		if err != nil {
			fmt.Println(err)
			return 0
		}

		defer file.Close()
		reader = bufio.NewReader(file)
	}

	scanner := bufio.NewScanner(reader)
	return uint64(countAllLines(scanner, substr))
}

func countAllLines(scanner *bufio.Scanner, substr string) int {
	count := 0
	for scanner.Scan() {
		count += strings.Count(scanner.Text(), substr)
	}
	return count
}
