package statsite

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func Assert(t *testing.T, expected, obtained interface{}) {
	if expected != obtained {
		t.Fatalf("Expected value [%v] not equal to obtained value [%v]", expected, obtained)
	}
}

func TestMessage(t *testing.T) {
	key := "foo"
	val := "bar"
	typ := TYPE_KEY_VALUE

	m := message{key, val, typ}

	Assert(t, key, m.Key)
	Assert(t, val, m.Value)
	Assert(t, typ, m.Type)

	Assert(t, fmt.Sprintf(MESSAGE_FORMAT, key, val, typ), m.String())
}

func TestKeyValueMessage(t *testing.T) {
	key := "foo"
	val := "bar"

	m := NewKeyValue(key, val).(*message)

	Assert(t, key, m.Key)
	Assert(t, val, m.Value)
	Assert(t, TYPE_KEY_VALUE, m.Type)
}

func TestKeyValueIntMessage(t *testing.T) {
	key := "foo"
	val := "10"

	m := NewKeyValueInt(key, 10).(*message)

	Assert(t, key, m.Key)
	Assert(t, val, m.Value)
	Assert(t, TYPE_KEY_VALUE, m.Type)
}

func TestGaugeMessage(t *testing.T) {
	key := "foo"
	val := "bar"

	m := NewGauge(key, val).(*message)

	Assert(t, key, m.Key)
	Assert(t, val, m.Value)
	Assert(t, TYPE_GAUGE, m.Type)
}

func TestTimerDurationMessage(t *testing.T) {
	key := "foo"
	dur := time.Minute

	m := NewTimerDuration(key, dur).(*message)

	Assert(t, key, m.Key)
	Assert(t, "60000", m.Value)
	Assert(t, TYPE_TIMER, m.Type)

}

func TestTimerMessage(t *testing.T) {
	key := "foo"
	tim1, _ := time.Parse(time.Stamp, "Jan 02 15:04:05")
	tim2, _ := time.Parse(time.Stamp, "Jan 02 15:04:06")

	m := NewTimer(key, tim1, tim2).(*message)

	Assert(t, key, m.Key)
	Assert(t, "1000", m.Value)
	Assert(t, TYPE_TIMER, m.Type)
}

func TestCounterMessage(t *testing.T) {
	key := "foo"
	val := 10

	m := NewCounter(key, val).(*message)

	Assert(t, key, m.Key)
	Assert(t, strconv.FormatInt(10, 10), m.Value)
	Assert(t, TYPE_COUNTER, m.Type)
}

func TestCounter64Message(t *testing.T) {
	key := "foo"

	m := NewCounter64(key, 10).(*message)

	Assert(t, key, m.Key)
	Assert(t, strconv.FormatInt(10, 10), m.Value)
	Assert(t, TYPE_COUNTER, m.Type)
}

func TestSetMessage(t *testing.T) {
	key := "foo"
	val := "bar"

	m := NewSet(key, val).(*message)

	Assert(t, key, m.Key)
	Assert(t, val, m.Value)
	Assert(t, TYPE_SET, m.Type)
}

func TestSetIntMessage(t *testing.T) {
	key := "foo"
	val := 10

	m := NewSetInt(key, val).(*message)

	Assert(t, key, m.Key)
	Assert(t, strconv.FormatInt(10, 10), m.Value)
	Assert(t, TYPE_SET, m.Type)
}
