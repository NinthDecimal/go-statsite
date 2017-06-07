package statsite

import (
	"log"
	"sync"
	"time"
)

// enabled controls whether to enable StatsiteMetrics
var enabled = false
var l sync.Mutex

// ErrorWaitTime represents the Time to wait on error
var ErrorWaitTime = time.Duration(10 * time.Second)

// ShutdownTimeout represents how long each waitgroup is given to shutdown
// cleanly
var ShutdownTimeout = time.Duration(10 * time.Second)

// Metric Prefix
var metricPrefix string

var statQueue chan Message
var flushWG sync.WaitGroup

// Initialize creates a new statsite client and starts the flusher
func Initialize(hostname string, prefix string) {
	client := NewClient(hostname)
	log.Printf("Starting stats collector [%s] on [%s]\n", prefix, hostname)
	InitializeWithClient(prefix, client)
}

// InitializeWithClient creates takes a statsite client and starts the flusher
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
	defer client.Close()

	if err != nil {
		goto Wait
	}

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

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		wg.Wait()
	}()
	select {
	case <-finished:
		return true
	case <-time.After(timeout):
		return false
	}
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
	waitTimeout(&publishWG, ShutdownTimeout)
	// Close the statQueue signaling the flusher to flush all enququed metrics
	// and exit
	close(statQueue)
	// Wait for the flusher to flush all enqueue metrics
	waitTimeout(&flushWG, ShutdownTimeout)
	// Disable Flushing
	disable()
}
