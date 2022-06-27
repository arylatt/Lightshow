package lightshow

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// EventTypeFade is an event which fades from a starting to an ending state linearly over a period of time
const EventTypeFade = iota

var (
	// ErrEventNoStates indicates that there is no start and end state associated with the event
	ErrEventNoStates = errors.New("start and end states are both off")

	// ErrEventStartAfterEnd indicates that the start time of the event is after the end time of the event
	ErrEventStartAfterEnd = errors.New("starts after it ends")

	// ErrEventNoLights indicates that there are no lights associated with the event
	ErrEventNoLights = errors.New("impacts no lights")

	// EventFrequency is the number of seconds to wait between steps.
	// By default it is 0.08s, because the hub recommends 12.5Hz
	EventFrequency = (time.Millisecond * 80)
)

// Event represents an action to take over a period of time
type Event struct {
	Type       int     `json:"type"`
	StartState State   `json:"startState,omitempty"`
	EndState   State   `json:"endState,omitempty"`
	StartTime  float64 `json:"startTime"`
	EndTime    float64 `json:"endTime"`
	Lights     []int   `json:"lights"`
}

// Validate runs some basic checks to ensure an event is valid
func (e *Event) Validate() error {
	errFmt := "event is invalid: %w"
	t := time.Now()

	if e.StartState == StateOff && e.EndState == StateOff {
		return fmt.Errorf(errFmt, ErrEventNoStates)
	}

	if t.Add(time.Second * time.Duration(e.StartTime)).After(t.Add(time.Second * time.Duration(e.EndTime))) {
		return fmt.Errorf(errFmt, ErrEventStartAfterEnd)
	}

	if len(e.Lights) == 0 {
		return fmt.Errorf(errFmt, ErrEventNoLights)
	}

	return nil
}

// CollidesWith returns true if there is a collision with the supplied event
func (e *Event) CollidesWith(other Event) bool {
	t := time.Now()

	if t.Add(time.Second * time.Duration(other.StartTime)).After(t.Add(time.Second * time.Duration(e.EndTime))) {
		return false
	}

	if t.Add(time.Second * time.Duration(other.StartTime)).Equal(t.Add(time.Second * time.Duration(e.EndTime))) {
		return false
	}

	if t.Add(time.Second * time.Duration(e.StartTime)).After(t.Add(time.Second * time.Duration(other.EndTime))) {
		return false
	}

	if t.Add(time.Second * time.Duration(e.StartTime)).Equal(t.Add(time.Second * time.Duration(other.EndTime))) {
		return false
	}

	for _, myLight := range e.Lights {
		for _, otherLight := range other.Lights {
			if myLight == otherLight {
				return true
			}
		}
	}

	return false
}

// run will run the sequence of steps in an event
func (e *Event) run(ctx context.Context, l *Lightshow) {
	t := time.Now()

	defer func() {
		fmt.Printf("Elapsed for event: %f\r\n", time.Since(t).Seconds())
	}()

	ctx, done := context.WithCancel(ctx)
	defer done()

	steps := []State{e.EndState}

	if e.StartTime != e.EndTime {
		stepCount := (time.Second*time.Duration(e.EndTime) - time.Second*time.Duration(e.StartTime)) / EventFrequency
		steps = e.StartState.StepsTo(int(stepCount), e.EndState)
	}

	go steps[0].run(ctx, done, l, e.Lights, steps[1:])

	<-ctx.Done()
}
