package lightshow

import (
	"context"
	"testing"
)

func TestEventRun(t *testing.T) {
	ev := Event{
		StartState: State{[3]int{255, 255, 255}, 1},
		EndState:   StateOff,
		StartTime:  0,
		EndTime:    10,
		Lights:     []int{1},
	}

	ev.run(context.Background(), nil)
}
