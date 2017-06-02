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
var flushWG sync.WaitGroup

func Initialize(hostname string, prefix string) {
	enabled = true
	metricPrefix = prefix
	statQueue = make(chan Message, 4096)
	go flush(hostname)
	log.Printf("Starting stats collector [%s] on [%s]\n", prefix, hostname)
}

func flush(hostname string) {
	flushWG.Add(1)
	defer flushWG.Done()
	if !enabled {
		return
	}

	var client Client
	var err error

Connect:
	// Initializes a statsite client based on the toml config file
	// Returns a statsite.Client and an error
	client, err = NewClient(hostname)
	if err != nil {
		log.Println("Error connecting to statsite")
		goto Wait
	}

	for {
		msg, more := <-statQueue
		if more {
			// More stats to receive
			err := client.Emit(msg)
			if err != nil {
				log.Println("Failed to write to statsite. Error: ", err)
			}
		} else {
			// statQueue channel closed and all stats received, exiting
			return
		}
	}

Wait:
	sleep := time.After(ErrorWaitTime)
	for {
		select {
		case <-statQueue:
			// Flush any messages sent before re-connecting
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
	// No more metrics, disable the flusher and return
	enabled = false
	close(statQueue)
	flushWG.Wait()

}
