package main

import (
	"flag"
	"github.com/kiip/go-statsite"
	"math/rand"
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

	statsite.Initialize(statsiteHost, "test")

	for i := 0; i < iterations; i++ {
		InstrumentedMethod()
	}
}

func InstrumentedMethod() {
	timer := statsite.Timer("test")
	defer timer.Emit()
	counter := statsite.CounterAt("test", rand.Intn(10))
	defer counter.Emit()

	<-time.After(time.Millisecond * time.Duration(rand.Intn(1000)))
}
