package statsite

import (
	"fmt"
	"time"
)

type Metric interface {
	Emit()
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
	if !enabled {
		return
	}

	go func(key string, start, end time.Time) {
		select {
		case statQueue <- NewTimer(
			fmt.Sprintf("%s.%s", metricPrefix, key),
			start,
			end,
		):
		default:
		}
	}(t.key, t.start, time.Now())
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
	if !enabled {
		return
	}

	go func(key string, count int) {
		select {
		case statQueue <- NewCounter(
			fmt.Sprintf("%s.%s", metricPrefix, key),
			count,
		):
		default:
		}
	}(t.key, t.count)
}

type timerCounter struct {
	counter counter
	timer   timer
}

func TimerCounter(key string) *timerCounter {
	return &timerCounter{
		timer{time.Now(), key},
		counter{key, 0},
	}
}

func (t *timerCounter) Emit() {
	if !enabled {
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
	if !enabled {
		return
	}

	go func(key string, value string) {
		select {
		case statQueue <- NewKeyValue(
			fmt.Sprintf("%s.%s", metricPrefix, key),
			value,
		):
		default:
		}
	}(t.key, t.value)
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
	if !enabled {
		return
	}

	go func(key string, value int) {
		select {
		case statQueue <- NewGauge(
			fmt.Sprintf("%s.%s", metricPrefix, key),
			value,
		):
		default:
		}
	}(t.key, t.value)
}
