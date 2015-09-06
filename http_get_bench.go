// Package main
/*
* http_get_bench - provides some benchmarks around an http get request
* my first foray in to go lang, and so far, very pleasant :)
*
* Author: Paul Davis
* Date: 2015.Feb.01
*
 */
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// Site List
type Site struct {
	Name string
	URL  string
}

// Metrics counters
type Metrics struct {
	byteCount, sampleCount             int
	latency, firstPayloadLatency       int64
	avgLatency, minLatency, maxLatency int64
	jitter, interPayloadLatency        int64
	timings                            []int64
}

var b bytes.Buffer

func count(name, url string) {
	start := time.Now()
	fmt.Fprintf(os.Stdout, "Started at "+start.String()+"\n")

	// init to get inter payload delay
	lastPkt := time.Now()

	r, err := http.Get(url)
	if err != nil {
		fmt.Printf("%s: %s", name, err)
		return
	}
	reqEndTime := time.Now()

	var n int64
	var firstPkt int64
	var timings []int64

	for i := 0; ; i++ {
		// TODO really should account for overhead... later
		chunks, err := io.CopyN(ioutil.Discard, r.Body, 1500)

		n = n + chunks
		if err == nil {
			// record time to read first payload
			if i == 0 {
				firstPkt = time.Since(lastPkt).Nanoseconds()
				latency := time.Since(reqEndTime).Nanoseconds()
				latencyMs := time.Since(reqEndTime).String()
				fmt.Printf("%s: Request start timestamp [%d]\n", name, firstPkt)
				fmt.Printf("%s: Response latency to first %d bytes from request end time(ns): [%d]\n", name, n, latency)
				fmt.Printf("%s: First packet latency in ms: %s\n", name, latencyMs)
				timings = append(timings, latency)

			} else {
				curPkt := time.Since(start).Nanoseconds()
				interPkt := time.Since(lastPkt).Nanoseconds()
				timings = append(timings, interPkt)

				// only print a limited number
				if (i % 10) == 0 {
					fmt.Printf("[%16d] bytes at timestamp [%13d] w/ latency (ns) [%12d]\n", n, curPkt, interPkt)
				}
			}

			//			fmt.Printf("Variance from last packet: [%d]\n", time.Since(lastPkt).Nanoseconds())
			lastPkt = time.Now()
		}

		if err != nil {
			fmt.Printf("%s: %s Received, exiting!\n", name, err)
			break
		}
	}

	r.Body.Close()
	l := len(timings)
	var min, max, avg int64
	var c int64
	for i := 0; i < l; i++ {
		c = c + timings[i]
		if i == 0 {
			min = timings[i]
		}
		if timings[i] > 0 && timings[i] < min {
			min = timings[i]
		}
		if timings[i] > 0 && timings[i] > max {
			max = timings[i]
		}
		avg = c / int64(l)

		// only need to see the last
		if i == (l - 1) {
			fmt.Printf("Inter Packet Delay Stats in ns: [min: %d | avg: %d | max: %d]\n", min, avg, max)
		}
	}

	fmt.Printf("Site: %s delivered %d bytes in:[%.2fs]\n", name, n, time.Since(start).Seconds())

	fmt.Printf("In milliseconds: %s \n", time.Since(start).String())

}

func do(f func(Site)) {
	// populate our struct w/ some data
	//site := Site{"Google", "http://www.google.com/"}
	site := Site{"Xfinity", "http://xfinity.comcast.net/"}

	f(site)

}

func main() {

	do(func(site Site) {
		fmt.Printf("%v\n", site)
		count(site.Name, site.URL)
	})

	fmt.Fprintf(os.Stdout, "Finished at "+time.Now().String()+"\n")
}
