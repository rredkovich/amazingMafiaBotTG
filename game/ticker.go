package game

import (
	"sync"
	"time"
)

type Ticker struct {
	mux                     sync.Mutex
	value                   int32
	tickStep                int32
	toAlarmValue            int32
	lastBeforeAlarmValue    int32
	beforeAlarmNotification int32
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
		t.lastBeforeAlarmValue -= t.tickStep
	}
	time.Sleep(time.Duration(t.tickStep) * time.Second)
}

func (t *Ticker) RaiseAlarm(seconds uint32) {
	t.mux.Lock()
	defer t.mux.Unlock()

	t.toAlarmValue = int32(seconds) // could be negative if tickStep != 1
	t.lastBeforeAlarmValue = int32(seconds)
	t.beforeAlarmNotification = 30 // default notification time is 30 seconds
}

// PostponeAlarm moves alarm to seconds ahead
func (t *Ticker) PostponeAlarm(seconds uint) {
	t.mux.Lock()
	defer t.mux.Unlock()

	t.toAlarmValue += int32(seconds) // could be negative if tickStep != 1
	t.lastBeforeAlarmValue += int32(seconds)
}

// Time to notifify for soon alarm
func (t *Ticker) AlarmIsSoon() bool {
	t.mux.Lock()
	defer t.mux.Unlock()

	return t.lastBeforeAlarmValue == t.beforeAlarmNotification

}

func (t *Ticker) GetValue() uint32 {
	t.mux.Lock()
	defer t.mux.Unlock()

	return uint32(t.value)
}
