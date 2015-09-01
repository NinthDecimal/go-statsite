package statsite

import (
	"fmt"
	"strconv"
	"time"
)

const (
	MESSAGE_FORMAT = "%v:%v|%v\n"

	TYPE_KEY_VALUE = MessageType("kv") // - Simple Key/Value
	TYPE_GAUGE     = MessageType("g")  // - Same as kv, compatibility with statsd gauges
	TYPE_TIMER     = MessageType("ms") // - Timer
	TYPE_COUNTER   = MessageType("c")  // - Counter
	TYPE_SET       = MessageType("s")  // - Unique Set
)

type MessageType string

type Message interface {
	String() string
}

type message struct {
	Key   string
	Value string
	Type  MessageType
}

func (m message) String() string {
	return fmt.Sprintf(MESSAGE_FORMAT, m.Key, m.Value, m.Type)
}

func NewKeyValue(key, value string) Message {
	return &message{
		Key:   key,
		Value: value,
		Type:  TYPE_KEY_VALUE,
	}
}

func NewGauge(key string, value int) Message {
	return &message{
		Key:   key,
		Value: strconv.FormatInt(int64(value), 10),
		Type:  TYPE_GAUGE,
	}
}

func NewTimer(key string, start, end time.Time) Message {
	return NewTimerDuration(key, end.Sub(start))
}

func NewTimerNow(key string, previous time.Time) Message {
	return NewTimerDuration(key, time.Now().Sub(previous))
}

func NewTimerDuration(key string, duration time.Duration) Message {
	value := int64(duration / time.Millisecond)
	return &message{
		Key:   key,
		Value: strconv.FormatInt(value, 10),
		Type:  TYPE_TIMER,
	}
}

func NewCounter(key string, value int) Message {
	return NewCounter64(key, int64(value))
}

func NewCounter64(key string, value int64) Message {
	return &message{
		Key:   key,
		Value: strconv.FormatInt(value, 10),
		Type:  TYPE_COUNTER,
	}
}

func NewSet(key, value string) Message {
	return &message{
		Key:   key,
		Value: value,
		Type:  TYPE_SET,
	}
}

func NewSetInt(key string, value int) Message {
	return NewSet(key, strconv.FormatInt(int64(value), 10))
}
