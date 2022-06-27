package lightshow

import (
	"errors"
	"fmt"
	"sort"
)

var (
	// ErrMapNoEvents indicates that the map contains no events
	ErrMapNoEvents = errors.New("no events")

	// ErrMapEventCollides indicates that there are overlapping/colliding events within the map
	ErrMapEventCollides = errors.New("event collides with another event")
)

// Map represents a playable map
type Map struct {
	Name   string  `json:"name"`
	Events []Event `json:"events"`
}

// SortEvents ensures that the events list is correctly sorted for execution
func (m *Map) SortEvents() {
	sort.SliceStable(m.Events, func(i, j int) bool {
		return m.Events[i].StartTime < m.Events[j].StartTime
	})
}

// Validate checks to see if there are any errors with the map
func (m *Map) Validate() error {
	errFmt := "map is invalid: %w"

	if len(m.Events) == 0 {
		return fmt.Errorf(errFmt, ErrMapNoEvents)
	}

	for i, ev1 := range m.Events {
		if err := ev1.Validate(); err != nil {
			return fmt.Errorf(errFmt, err)
		}

		for j, ev2 := range m.Events {
			if i == j {
				continue
			}

			if ev1.CollidesWith(ev2) {
				return fmt.Errorf("map is invalid, event1=%v; event2=%v: %w", ev1, ev2, ErrMapEventCollides)
			}
		}
	}

	return nil
}
