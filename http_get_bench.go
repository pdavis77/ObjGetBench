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
				fmt.Printf("%s: Request start to First %d bytes in [%d]\n", name, n, firstPkt)
				fmt.Printf("%s: Response latency to first %d bytes from request end time: [%d]\n", name, n, latency)
				timings = append(timings, latency)

			} else {
				curPkt := time.Since(start).Nanoseconds()
				interPkt := time.Since(lastPkt).Nanoseconds()
				timings = append(timings, interPkt)

				fmt.Printf("[%16d] bytes in [%16d] w/ latency [%16d]\n", n, curPkt, interPkt)

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

		fmt.Printf("min: %d | avg: %d | max: %d\n", min, avg, max)

	}
	fmt.Printf("%s %d [%.2fs]\n", name, n, time.Since(start).Seconds())

	fmt.Printf("It only took me: %s \n", time.Since(start).String())

}

func do(f func(Site)) {
	// populate our struct w/ some data
	site := Site{"Google", "http://www.google.com/"}
	//site := Site{"Xfinity", "http://xfinity.comcast.net/"}

	f(site)

}

func main() {

	do(func(site Site) {
		fmt.Printf("%v\n", site)
		count(site.Name, site.URL)
	})

	fmt.Fprintf(os.Stdout, "Finished at "+time.Now().String()+"\n")
}
