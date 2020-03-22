package game

import (
	"sync"
	"time"
)

type Ticker struct {
	mux                  sync.Mutex
	value                uint32
	tickStep             uint32
	toAlarmValue         int32
	lastBeforeAlarmValue int32
}

func (t *Ticker) Alarm() bool {
	t.mux.Lock()
	defer t.mux.Unlock()

	return t.lastBeforeAlarmValue <= 0
}

func (t *Ticker) Tick() {
	t.mux.Lock()
	defer t.mux.Unlock()

	t.value += t.tickStep
	if t.toAlarmValue > 0 {
		t.lastBeforeAlarmValue -= int32(t.tickStep)
	}
	time.Sleep(time.Duration(t.tickStep) * time.Second)
}

func (t *Ticker) RaiseAlarm(seconds uint32) {
	t.mux.Lock()
	defer t.mux.Unlock()

	t.toAlarmValue = int32(seconds) // could be negative if tickStep != 1
	t.lastBeforeAlarmValue = int32(seconds)
}

// PostponeAlarm moves alarm to seconds ahead
func (t *Ticker) PostponeAlarm(seconds uint) {
	t.mux.Lock()
	defer t.mux.Unlock()

	t.toAlarmValue += int32(seconds) // could be negative if tickStep != 1
	t.lastBeforeAlarmValue += int32(seconds)
}

func (t *Ticker) GetValue() uint32 {
	t.mux.Lock()
	defer t.mux.Unlock()

	return t.value
}
