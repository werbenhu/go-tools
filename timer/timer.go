package timer

import (
	"sync"
	"time"

	"git.aimore.com/golang/timer/timewheel"
)

var tw *timewheel.TimeWheel
var timers map[string]*timewheel.Timer
var timersMu sync.Mutex

func init() {
	tw = timewheel.New(time.Second, 10)
	timers = make(map[string]*timewheel.Timer)
}

func Cancel(id string) {
	timersMu.Lock()
	defer timersMu.Unlock()
	if timer, ok := timers[id]; ok {
		timer.Stop()
	}
	delete(timers, id)
}

func Start(seconds int, fn func(id string)) (string, error) {
	timersMu.Lock()
	defer timersMu.Unlock()
	t := tw.AfterFunc(time.Duration(seconds)*time.Second, func(id string) {
		delete(timers, id)
		fn(id)
	})
	timers[t.Id()] = t
	return t.Id(), nil
}

func Count() int {
	timersMu.Lock()
	defer timersMu.Unlock()
	return len(timers)
}

func Schedule(seconds int, fn func(id string)) (string, error) {
	timersMu.Lock()
	defer timersMu.Unlock()
	schedule := &DefaultScheduler{Interval: time.Second * time.Duration(seconds)}
	t := tw.ScheduleFunc(schedule, fn)
	timers[t.Id()] = t
	return t.Id(), nil
}

func Run() {
	tw.Start()
}
