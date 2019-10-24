package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

var wg sync.WaitGroup

func main()  {
	var (
		length   int
		limit    int
		subLen   int
		diff     int
		url      string
		filename string
		err      error
		fd       *os.File
		chunk int
	)
	url = "http://192.168.1.201:8003/api/client/downloader"
	limit = 10
	chunk = 2048

	res, _ := http.Head(url)
	defer res.Body.Close()
	maps := res.Header
	log.Println(maps)
	length, _ = strconv.Atoi(maps["Content-Length"][0])
	log.Println("length:", length)
	filename = strings.Split(maps["Content-Disposition"][0], "=")[1]
	log.Printf("filename:%s\n", filename)

	fd, err = os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Printf("open file error:%v\n", err)
		return
	}
	defer fd.Close()

	subLen = length / limit
	diff = length % limit //the remaining for the last request.

	fmt.Println(subLen, diff)
	for i := 0; i < limit; i++ {
		wg.Add(1)
		min := subLen * i
		max := subLen * (i+1)

		if i == limit -1 {
			max += diff
		}

		go func(min, max int) {
			defer wg.Done()
			var (
				seek int64
				total int
				reqs int //request times of this goroutine
				diff int
				nRead int64
				gfd       *os.File
				//err  error
			)
			gfd, err = os.OpenFile(filename, os.O_RDWR, 0755)
			if err != nil {
				log.Printf("open file error:%v\n", err)
				return
			}
			seek = int64(min)
			client := &http.Client{}

			total = max - min
			if total <= chunk {
				reqs = 1
			}
			reqs = total / chunk
			diff = total % chunk

			for n := 0; n < reqs; n++ {
				start := int(seek)
				end := start + chunk
				if n == reqs -1 {
					end += diff
				}

				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					log.Println("NewRequest error:", err)
					return
				}

				rangeHeader := "bytes=" + strconv.Itoa(start) + "-" + strconv.Itoa(end-1)
				req.Header.Add("Range", rangeHeader)
				resp, err := client.Do(req)
				if err != nil {
					log.Println("req error:", err)
					return
				}

				_, err = gfd.Seek(seek, 0)
				if err != nil {
					log.Println("seek error:", err)
					return
				}
				nRead, err = io.Copy(gfd, resp.Body)
				if err != nil {
					log.Println("Copy error:", err)
					return
				}
				_ = resp.Body.Close()
				seek += nRead
				log.Println("new seek:", seek)
			}
		}(min, max)
	}
	wg.Wait()
}
