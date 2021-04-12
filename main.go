package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"gosuri/uiprogress"
	"log"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

var count uint64
var countDiff uint64
var count10 uint64
var count20 uint64
var count30 uint64
var count85 uint64
var maxDuration time.Duration

func main() {
	file := flag.String("f", "", "file to process")
	host1 := flag.String("h1", "", "host1")
	host2 := flag.String("h2", "", "host2")
	xclient := flag.String("xclient", "test", "xclient")
	th := flag.Int64("th", 1, "threads")
	token1 := flag.String("token1", "", "atp token host1")
	token2 := flag.String("token2", "", "atp token host2")
	flag.Parse()

	//verify mandatory parameters
	if *file == "" || *host1 == "" || *host2 == "" {
		usage()
	}

	//define max number of thread in 1000
	if *th > 10 {
		*th = 10
	}

	//star bar
	uiprogress.Start()
	bar := uiprogress.AddBar(countLines(*file))

	//begin the process
	f, err := os.Open(*file)
	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}
	fileScanner := bufio.NewScanner(f)
	var wg sync.WaitGroup
	var ops int64

	for fileScanner.Scan() {
		for ops > *th {
			//wait a free thread
			time.Sleep(time.Millisecond * 100)
		}
		wg.Add(1)
		atomic.AddInt64(&ops, 1)
		url := fileScanner.Text()
		go worker(&ops, &wg, url, *host1, *host2, *xclient, *token1, *token2)
		bar.Incr()
	}
	wg.Wait()
	uiprogress.Stop()
	fmt.Printf("Cantidad de request %d\n", count)
	fmt.Printf("Cantidad de request diferentes %d\n", countDiff)
	fmt.Printf("Cantidad de request > 10ms %d\n", count10)
	fmt.Printf("Cantidad de request > 20ms %d\n", count20)
	fmt.Printf("Cantidad de request > 30ms %d\n", count30)
	fmt.Printf("Cantidad de request > 85ms %d\n", count85)
	fmt.Printf("Max duration %d\n", int64(maxDuration/time.Millisecond))

}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(-1)
}

func worker(id *int64, wg *sync.WaitGroup, url string, host1 string, host2 string, xclient string, token1 string, token2 string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered in %s error %v\n\n", url, r)
		}
		atomic.AddInt64(id, -1)
		wg.Done()
	}()

	restClient1 := NewRestClient(xclient, token1)
	restClient2 := NewRestClient(xclient, token2)

	var startTime = time.Now()
	bodyServ1, _ := restClient1.get(fmt.Sprintf("%s%s", host1, url))
	var duration = time.Since(startTime)

	bodyServ2, _ := restClient2.get(fmt.Sprintf("%s%s", host2, url))

	if duration > maxDuration {
		maxDuration = duration
	}
	dur := int64(duration / time.Millisecond)
	if dur > 85 {
		atomic.AddUint64(&count85, 1)
		//fmt.Printf("%d %s%s %v\n\n", dur, host1, url, bodyServ1)
	} else if dur > 30 {
		atomic.AddUint64(&count30, 1)
	} else if dur > 20 {
		atomic.AddUint64(&count20, 1)
	} else if dur > 10 {
		atomic.AddUint64(&count10, 1)
	}

	if !areEqualJSON(bodyServ1, bodyServ2) {
		fmt.Printf("1 %s %v\n\n", url, bodyServ1)
		fmt.Printf("2 %s %v\n\n", url, bodyServ2)
		atomic.AddUint64(&countDiff, 1)
	}
	atomic.AddUint64(&count, 1)
}

func areEqualJSON(s1 string, s2 string) bool {
	var o1 interface{}
	var o2 interface{}
	var err error
	err = json.Unmarshal([]byte(s1), &o1)
	if err != nil {
		return false //, fmt.Errorf("Error mashalling string 1 :: %s", err.Error())
	}
	err = json.Unmarshal([]byte(s2), &o2)
	if err != nil {
		return false //, fmt.Errorf("Error mashalling string 2 :: %s", err.Error())
	}
	return reflect.DeepEqual(o1, o2)
}

func countLines(file string) int {
	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}
	fileScanner := bufio.NewScanner(f)
	lineCount := 0
	for fileScanner.Scan() {
		lineCount++
	}
	f.Close()
	return lineCount
}
