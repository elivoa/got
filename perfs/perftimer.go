package perfs

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

// TODO check utils.PerfTimer

// Timer uses to mearsure time of operation
type Timer struct {
	time.Time
	lap  time.Time
	Laps *[]LapInfo // 记录圈数
	name string
	id   string

	// flags
	disableLaps bool
}

type LapInfo struct {
	Name      string        `consumer:"n"`
	Duration  time.Duration `consumer:"d"`
	DurString string        `consumer:"s"`
}

var isLog = os.Getenv("GO_OUTPUT_PERFTIME_LOG") != "OFF"

// NewPerfTimer creates a new perf time object
func NewPerfTimer(name string) (now *Timer) {
	id := fmt.Sprintf("%08X", rand.Uint32())
	return NewPerfTimerWithID(id, name)
}

// NewPerfTimerWithID creates a new perf time object
func NewPerfTimerWithID(id string, name string) (now *Timer) {
	timenow := time.Now()
	tempTime := Timer{Time: timenow, lap: timenow, name: name, id: id}
	if isLog {
		log.Printf("%s Timing ❯ %v", tempTime.id, name)
	}
	return &tempTime
}

func (t *Timer) DisableLaps() (now *Timer) {
	t.disableLaps = true
	return t
}

func (t *Timer) EnableLaps() (now *Timer) {
	t.disableLaps = false
	return t
}

func (t *Timer) LapDuration() (duration time.Duration) {
	if t.disableLaps {
		return 0
	}
	return time.Since(t.lap)
}

func (t *Timer) ResetLap() {
	if !t.disableLaps {
		t.lap = time.Now()
	}
}

func (t *Timer) Lap() (duration time.Duration) {
	if t.disableLaps {
		return 0
	}
	return t.LapInfo("")
}

func (t *Timer) LapInfo(extraInfo string) (duration time.Duration) {
	if t.disableLaps {
		return
	}
	if isLog {
		duration = t.LapDuration()
		t.ResetLap()
		t.addLap(extraInfo, duration)
		log.Printf("%s ❯ %v ❯ %v  + %v", t.id, t.name, extraInfo, duration)
	}
	return
}

func (t *Timer) addLap(extraInfo string, duration time.Duration) {
	if t.Laps == nil {
		laps := []LapInfo{}
		t.Laps = &laps
	}
	// safty check
	if len(*t.Laps) > 1000 {
		*t.Laps = (*t.Laps)[len(*t.Laps)-1000 : len(*t.Laps)]
	}
	// add to laps
	*t.Laps = append(*t.Laps, LapInfo{
		Name: extraInfo, Duration: duration, DurString: fmt.Sprintf("%v", duration),
	})
}

// SoFar logs time spent.
func (t *Timer) SoFar() (duration time.Duration) {
	return t.SoFarInfo("")
}

// SoFarInfo logs time spent with information
func (t *Timer) SoFarInfo(extraInfo string) (duration time.Duration) {
	if isLog {
		duration = t.Duration()
		log.Printf("%s ❯ %v ❯ %v + total %v", t.id, t.name, extraInfo, duration)
	}
	return
}

// ID returns internal ID for more logging purpose.
func (t *Timer) ID() string {
	return t.id
}

// Duration returns duration so far
func (t *Timer) Duration() (duration time.Duration) {
	return time.Since(t.Time)
}
