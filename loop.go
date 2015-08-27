package statsite

import (
	"log"
	"time"
)

// Whether to enable StatSiteMetrics
var enabled bool = false

// Time to wait on error
var ErrorWaitTime time.Duration = time.Duration(10 * time.Second)

// Metric Prefix
var metricPrefix string

var statQueue chan Message

func Initialize(hostname string, prefix string) {
	enabled = true
	metricPrefix = prefix
	statQueue = make(chan Message, 4096)
	go flush(hostname)
	log.Printf("Starting stats collector [%s] on [%s]\n", prefix, hostname)
}

func flush(hostname string) {
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
		select {
		case msg := <-statQueue:
			if err := client.Emit(msg); err != nil {
				log.Println("Failed to write to statsite. Error: ", err)
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
		case <-sleep:
			goto Connect
		}
	}
}
