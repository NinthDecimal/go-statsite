package statsite

import (
	"log"
	"sync"
	"time"
)

// Whether to enable StatSiteMetrics
var enabled bool = false
var l sync.Mutex

// Time to wait on error
var ErrorWaitTime time.Duration = time.Duration(10 * time.Second)

// Metric Prefix
var metricPrefix string

var statQueue chan Message
var flushWG sync.WaitGroup

func Initialize(hostname string, prefix string) {
	client := NewClient(hostname)
	log.Printf("Starting stats collector [%s] on [%s]\n", prefix, hostname)
	InitializeWithClient(prefix, client)
}

func InitializeWithClient(prefix string, client Client) {
	enable()
	metricPrefix = prefix
	statQueue = make(chan Message, 4096)
	flushWG.Add(1)
	go flush(client)
}

func flush(client Client) {
	defer flushWG.Done()
	if !enabled {
		return
	}

	var err error

Connect:
	// Initializes a statsite client based on the toml config file
	// Returns a statsite.Client and an error
	// var client Client
	err = client.Connect()
	if err != nil {
		goto Wait
	}
	// defer client.Close()

	for {
		msg, more := <-statQueue
		if more {
			// More stats to receive
			err := client.Emit(msg)
			if err != nil {
				log.Println("Failed to write to statsite. Error: ", err)
				goto Wait
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

func enable() {
	l.Lock()
	enabled = true
	publishEnabled = true
	l.Unlock()
}

func disablePublish() {
	l.Lock()
	publishEnabled = false
	l.Unlock()
}

func disable() {
	l.Lock()
	enabled = false
	l.Unlock()
}

// Shutdown is used to cleanly shutdown go-statsite, flushing all metrics before
// exiting.
func Shutdown() {
	if !enabled {
		return
	}
	// Disable publishing new metrics
	disablePublish()
	// Wait for all in-flight metrics to be added to the statQueue
	publishWG.Wait()
	// Close the statQueue signaling the flusher to flush all enququed metrics
	// and exit
	close(statQueue)
	// Wait for the flusher to flush all enqueue metrics
	flushWG.Wait()
	// Disable Flushing
	disable()

}
