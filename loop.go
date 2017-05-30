package statsite

import (
	"log"
	"sync"
	"time"
)

// Whether to enable StatSiteMetrics
var enabled bool = false

// Time to wait on error
var ErrorWaitTime time.Duration = time.Duration(10 * time.Second)

// Metric Prefix
var metricPrefix string

var statQueue chan Message
var shutdown chan bool
var wg sync.WaitGroup

func Initialize(hostname string, prefix string) {
	enabled = true
	metricPrefix = prefix
	statQueue = make(chan Message, 4096)
	shutdown = make(chan bool, 1)
	go flush(hostname)
	log.Printf("Starting stats collector [%s] on [%s]\n", prefix, hostname)
}

func flush(hostname string) {
	wg.Add(1)
	defer wg.Done()
	if !enabled {
		return
	}

	var client Client
	var err error
	var exit = false

Connect:
	// Initializes a statsite client based on the toml config file
	// Returns a statsite.Client and an error
	client, err = NewClient(hostname)
	if err != nil {
		log.Println("Error connecting to statsite")
		goto Wait
	}

	for {
		select {
		case msg := <-statQueue:
			if err := client.Emit(msg); err != nil {
				log.Println("Failed to write to statsite. Error: ", err)
				if exit {
					// stop the flusher
					return
				}
				goto Wait
			}
		}
	}

Wait:
	sleep := time.After(ErrorWaitTime)
	for {
		select {
		case <-statQueue:
			// Flush any messages sent before re-connecting
		case <-shutdown:
			log.Println("Shutting down go-statsite flusher")
			exit = true
			goto Connect
		case <-sleep:
			goto Connect
		}
	}
}

// Shutdown is used to cleanly shutdown go-statsite, flushing all metrics before
// exiting.
func Shutdown() {
	if !enabled {
		return
	}
	// Send the shutdown signal to the flusher
	shutdown <- true
	for {
		select {
		case <-statQueue:
			// More metrics left to publish
		default:
			// No more metrics, disable the flusher and return
			enabled = false
			wg.Wait()
		}
	}
}
