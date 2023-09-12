package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	utils "fable/utils"
)

func main() {

	conns := flag.Int("conns", 1000, "")
	host := flag.String("host", "producer", "")
	flag.Parse()

	hosturl := fmt.Sprintf("http://%s:8080/log", *host)
	var wg sync.WaitGroup
	var successReq, failedReq, httpErr, totalReq int64

	exitch := make(chan error, 1)
	go utils.ShutDownHandler(exitch)

	log.Printf("starting load testing on hosturl=%s, rps=%d", hosturl, *conns)
	ticker := time.NewTicker(time.Second / time.Duration(*conns))
	var client = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        *conns, // Adjust the maximum idle connections as per your requirement
			MaxIdleConnsPerHost: *conns, // Adjust the maximum idle connections per host as per your requirement
		},
	}

	startTime := time.Now()
	i := int64(0)
LOOP:
	for {
		select {
		case <-exitch:
			ticker.Stop()
			wg.Wait()
			client.CloseIdleConnections()
			timeTaken := time.Since(startTime).Seconds()
			rps := float64(totalReq) / timeTaken
			fmt.Println("")
			fmt.Println(" ", totalReq)
			fmt.Println("-", httpErr)
			fmt.Println("request placed =", totalReq-httpErr)
			fmt.Println("sucess http req:", successReq)
			fmt.Println("failed http req:", failedReq)
			fmt.Println("total time live: ", timeTaken)
			fmt.Println("rps = ", rps)
			break LOOP
		case <-ticker.C:
			i += 1
			wg.Add(1)
			jsonIp, _ := json.Marshal(utils.LogInput{
				Id:        i,
				Event:     "login",
				UserId:    123,
				TimeStamp: 12354,
			})
			go func() {
				defer wg.Done()
				atomic.AddInt64(&totalReq, 1)
				resp, err := client.Post(hosturl, "application/json", bytes.NewBuffer(jsonIp))
				if err != nil {
					atomic.AddInt64(&httpErr, 1)
					return
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					atomic.AddInt64(&failedReq, 1)
					return
				}
				atomic.AddInt64(&successReq, 1)
			}()
		}
	}
}
