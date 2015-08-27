package main

import (
	"flag"
	"github.com/kiip/go-statsite"
	"time"
)

var statsiteHost string
var iterations int

func init() {
	flag.StringVar(&statsiteHost, "host", "localhost:8125", "Statsite host address")
	flag.IntVar(&iterations, "iterations", 10, "Iterations")
}

func main() {
	flag.Parse()

	client, err := statsite.NewClient(statsiteHost)
	if err != nil {
		panic(err)
	}

	for i := 0; i < iterations; i++ {
		start := time.Now()
		<-time.After(time.Second)
		end := time.Now()

		msg := statsite.NewTimer("test", start, end)

		if err := client.Emit(msg); err != nil {
			panic(err)
		}
	}
}
