package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type download struct {
	Url          string
	TargetPath   string
	TotalSegment int
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
func (d download) Run() error {
	req, err := d.getNewRequest("GET")
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	// fmt.Println(res.Status)
	if res.StatusCode > 299 {
		return errors.New(fmt.Sprintf("Can't process,response code is %v", res.StatusCode))
	}
	size, err := strconv.Atoi(res.Header.Get("Content-Length"))
	if err != nil {
		return err
	}
	var sections = make([][2]int, d.TotalSegment)
	each_size := size / d.TotalSegment
	fmt.Printf("each size of sections is %v\n", each_size)

	for i := range sections {
		if i == 0 {
			sections[i][0] = 0
		} else {
			sections[i][0] = sections[i-1][1] + 1
		}
		if i < d.TotalSegment-1 {
			sections[i][1] = sections[i][0] + each_size
		} else {
			sections[i][1] = size - 1
		}
	}
	log.Println(sections)

	var wg sync.WaitGroup
	for i, s := range sections {
		wg.Add(1)
		go func(i int, s [2]int) {
			defer wg.Done()
			err = d.downloadSection(i, s)
			if err != nil {
				panic(err)
			}
		}(i, s)
	}
	wg.Wait()

	return d.mergeFiles(sections)
}
func (d download) mergeFiles(sections [][2]int) error {
	file, err := os.OpenFile(d.TargetPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	for i := range sections {
		tmpFileName := fmt.Sprintf("section-%v.tmp", i)
		data, err := ioutil.ReadFile(tmpFileName)
		if err != nil {
			return err
		}
		n, err := file.Write(data)
		if err != nil {
			return err
		}
		err = os.Remove(tmpFileName)
		if err != nil {
			return err
		}
		fmt.Printf("%v bytes merged.\n", n)
	}
	return nil
}
func (d download) downloadSection(i int, c [2]int) error {
	req, err := d.getNewRequest("GET")
	if err != nil {
		return err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", c[0], c[1]))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode > 299 {
		return errors.New(fmt.Sprintf("Can't process, response is %v", res.StatusCode))
	}
	fmt.Printf("Downloaded %v bytes for section %v\n", res.Header.Get("Content-Length"), i)
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fmt.Sprintf("section-%v.tmp", i), data, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
func (d download) getNewRequest(method string) (*http.Request, error) {
	req, err := http.NewRequest(
		method,
		d.Url,
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "aliraee concurrent download manager v0.1")
	return req, nil

}
func main() {
	p := fmt.Println

	start_time := time.Now()

	urlptr := flag.String("url", "", "Download url")
	pathptr := flag.String("path", "",
		"Path to save downloding file with file name like: /path/someting.pdf")
	numptr := flag.Int("n", 5,
		"OPTIONAL: Number of go routin to download file sections")

	flag.Parse()
	p("Download address:", *urlptr)
	p("Path to save file:", *pathptr)
	p("Number of goroutine to donwload file:", *numptr)

	my_download := download{
		Url:          *urlptr,  //"https://bitcoin.org/bitcoin.pdf"
		TargetPath:   *pathptr, //"bitcoin-white-paper.pdf"
		TotalSegment: *numptr,  //5
	}

	err := my_download.Run()
	check(err)
	fmt.Printf("download is complete.\nTotaltime is %v\n", time.Now().Sub(start_time))

}
