package statsite

import (
	"fmt"
	"sync"
	"time"
)

var publishWG sync.WaitGroup
var publishEnabled bool = false

type Metric interface {
	Emit()
}

func publish(message Message) {
	defer publishWG.Done()
	statQueue <- message
}

// Timer Metric
// t := Timer(key)
// defer t.Emit()
type timer struct {
	start time.Time
	key   string
}

func Timer(key string) *timer {
	return &timer{time.Now(), key}
}

func (t *timer) Emit() {
	if !publishEnabled {
		return
	}
	timer := NewTimer(
		fmt.Sprintf("%s.%s", metricPrefix, t.key),
		t.start,
		time.Now(),
	)
	publishWG.Add(1)
	go publish(timer)
}

// Counter Metric
// t := Timer(key)
// defer t.Emit()
type counter struct {
	key   string
	count int
}

func Counter(key string) *counter {
	return &counter{key, 0}
}

func CounterAt(key string, i int) *counter {
	return &counter{key, i}
}

func (t *counter) Incr() {
	t.count += 1
}

func (t *counter) IncrBy(i int) {
	t.count += i
}

func (t *counter) Emit() {
	if !publishEnabled {
		return
	}

	counter := NewCounter(fmt.Sprintf("%s.%s", metricPrefix, t.key), t.count)
	publishWG.Add(1)
	go publish(counter)
}

type timerCounter struct {
	timer   *timer
	counter *counter
}

func TimerCounter(key string) *timerCounter {
	return &timerCounter{
		Timer(key),
		CounterAt(key, 1),
	}
}

func TimerCounterAt(key string, i int) *timerCounter {
	return &timerCounter{
		Timer(key),
		CounterAt(key, i),
	}
}

func (t *timerCounter) Incr() {
	t.counter.Incr()
}

func (t *timerCounter) IncrBy(i int) {
	t.counter.IncrBy(i)
}

func (t *timerCounter) Emit() {
	if !publishEnabled {
		return
	}

	t.counter.Emit()
	t.timer.Emit()
}

type keyvalue struct {
	key   string
	value string
}

func KeyValue(key string, value string) *keyvalue {
	return &keyvalue{key, value}
}

func (t *keyvalue) Emit() {
	if !publishEnabled {
		return
	}

	kv := NewKeyValue(fmt.Sprintf("%s.%s", metricPrefix, t.key), t.value)
	publishWG.Add(1)
	go publish(kv)
}

type gauge struct {
	key   string
	value int
}

func Gauge(key string) *gauge {
	return &gauge{key, 0}
}

func GaugeAt(key string, value int) *gauge {
	return &gauge{key, value}
}

func (t *gauge) Incr() {
	t.value += 1
}

func (t *gauge) IncrBy(i int) {
	t.value += i
}

func (t *gauge) Emit() {
	if !publishEnabled {
		return
	}

	guage := NewGauge(fmt.Sprintf("%s.%s", metricPrefix, t.key), t.value)
	publishWG.Add(1)
	go publish(guage)
}
