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
