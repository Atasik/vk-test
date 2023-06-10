package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type userSession struct {
	wg            *sync.WaitGroup
	totalCount    uint64
	goroutinesNum int
	client        http.Client
}

type job struct {
	path string
	file bool
}

func checkURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func countAllLines(scanner *bufio.Scanner, substr string) uint64 {
	count := 0
	for scanner.Scan() {
		count += strings.Count(scanner.Text(), substr)
	}
	return uint64(count)
}
func main() {
	// вынес все важные переменные во флаги
	var textToFind string
	var goroutinesNum int
	var timeout time.Duration

	flag.StringVar(&textToFind, "s", "Go", "text to find")
	flag.IntVar(&goroutinesNum, "n", 5, "number of gouroutines")
	flag.DurationVar(&timeout, "t", 2*time.Second, "timeout (seconds)")

	flag.Parse()

	u := &userSession{
		wg:            &sync.WaitGroup{},
		goroutinesNum: goroutinesNum,
		totalCount:    0,
		client: http.Client{
			Timeout: timeout,
		},
	}

	jobs := make(chan job)
	go u.countAll(jobs, textToFind)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		path := scanner.Text()
		switch path {
		case "":
			continue
		default:
			jobs <- job{path, checkURL(path)}
		}
	}

	u.wg.Wait()
	close(jobs)
	fmt.Printf("Total: %d", u.totalCount)
}

func (u *userSession) countAll(jobs chan job, textToFind string) {
	waitCh := make(chan struct{}, u.goroutinesNum) // канал, который блокирует добавление "лишних" горутин

	for job := range jobs {
		waitCh <- struct{}{}

		u.wg.Add(1)
		go u.incTotalCount(waitCh, job, textToFind)
	}

	close(waitCh)
}

func (u *userSession) incTotalCount(waitCh chan struct{}, job job, textToFind string) {
	defer u.wg.Done()

	qty := u.сountSource(job, textToFind)
	atomic.AddUint64(&u.totalCount, qty) // использую atomic, чтобы писать в total из горутины
	fmt.Printf("Count for %s: %d\n", job.path, qty)
	<-waitCh
}

func (u *userSession) сountSource(job job, textToFind string) uint64 {
	var reader io.Reader

	// проверяю, url или путь до файла
	if job.file {
		resp, err := u.client.Get(job.path)
		if err != nil {
			fmt.Println(err)
			return 0
		}
		defer resp.Body.Close()

		reader = bufio.NewReader(resp.Body)
	} else {
		file, err := os.Open(job.path)
		if err != nil {
			fmt.Println(err)
			return 0
		}
		defer file.Close()

		reader = bufio.NewReader(file)
	}

	scanner := bufio.NewScanner(reader)
	return countAllLines(scanner, textToFind)
}
