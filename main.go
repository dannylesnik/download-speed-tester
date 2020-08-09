package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"syscall"

	"github.com/dannylesnik/download-speed-tester/rand"

	"os"
	"os/signal"
	"sort"
	"strconv"
	"time"
)

func main() {

	requestURL := getEnvVar("HTTP_URL", "")
	_, err := url.ParseRequestURI(requestURL)
	if err != nil {
		log.Panic("Was not able to parse URL HTTP_URL as url")
	}

	numOfSamples, err := strconv.Atoi(getEnvVar("TOTAL_REQ", "25"))
	if err != nil {
		log.Panic("Was not able to parse TOTAL_REQ as integer.")
	}

	batchSize, err := strconv.Atoi(getEnvVar("MAX_PARAL_REQ", "5"))
	if err != nil {
		log.Panic("Was not able to parse MAX_PARAL_REQ as integer.")
	}
	numberOfBatches := numOfSamples / batchSize
	lastBatchSize := numOfSamples % batchSize
	data := make([]int64, 0)

	log.Println("Starting download with the following parameters:")
	log.Printf("URL (HTTP_URL): %s\n\n", requestURL)
	log.Printf("Total number of requests  (TOTAL_REQ): %d\n\n", numOfSamples)
	log.Printf("Maximum parallel requests (MAX_PARAL_REQ): %d\n\n\n", batchSize)

	ch := make(chan int64, batchSize)

	for j := 1; j <= numberOfBatches; j++ {
		results := runBatch(batchSize, ch, requestURL)
		data = append(data, results...)
	}

	results := runBatch(lastBatchSize, ch, requestURL)
	data = append(data, results...)
	close(ch)

	sort.Slice(data, func(i, j int) bool { return data[i] < data[j] })
	log.Println("******************** Reporting Results ***********************************************")
	log.Printf("\nShortest download was %d milliseconds\n\n", data[0])
	log.Printf("Longest download was %d milliseconds\n\n", data[len(data)-1])
	log.Printf("The Average is %d milliseconds\n\n", average(data))

	func() {
		interruptChan := make(chan os.Signal, 1)
		signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		// Block until we receive our signal.
		<-interruptChan

		log.Println("Shutting down")
		os.Exit(0)

	}()

}

func average(data []int64) int64 {
	length := len(data)
	var sum int64 = 0
	for i := 0; i < length; i++ {
		sum += data[i]
	}
	return sum / int64(length)
}

func runBatch(batchSize int, ch chan int64, url string) []int64 {
	results := make([]int64, batchSize)
	for i := 1; i <= batchSize; i++ {
		go download(url, ch)
	}

	for i := 0; i < batchSize; i++ {
		result := <-ch
		results[i] = result
		log.Printf("Download Completed.")
	}
	return results
}

func download(url string, ch chan<- int64) {
	startTime := time.Now()
	filename := rand.String(10)
	endTime := startTime
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error creating HTTP request: ", err.Error())
		ch <- 0
		return
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error making HTTP request: ", err.Error())
		ch <- 0
		return
	}

	if resp.StatusCode != 200 {
		log.Printf("Received status code %d. Skipping Download.", resp.StatusCode)
		ch <- 0
		return
	}
	const tenKB = 1024 * 10

	bytesRead := 0

	buf := make([]byte, tenKB)

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 666)
	if err != nil {
		ch <- 0
		panic(err)
	}

	var offset int64 = 0

	total := 0

	for {
		n, err := resp.Body.Read(buf)
		bytesRead += n

		_, ferr := f.WriteAt(buf, offset)
		total += n
		if ferr != nil {
			f.Close()
			panic(ferr)
		}
		offset += int64(n)
		if err == io.EOF {
			f.Close()
			endTime = time.Now()
			break
		}

	}
	err = os.Remove(filename)

	if err != nil {
		log.Println(err)
	}
	ch <- endTime.Sub(startTime).Milliseconds()
}

func getEnvVar(key string, defaultValue string) string {
	value, iSuccess := os.LookupEnv(key)

	if iSuccess {
		return value
	}
	return defaultValue
}
