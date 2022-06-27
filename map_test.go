package lightshow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapSortEvents(t *testing.T) {
	unsorted := []Event{{StartTime: 1}, {StartTime: 5}, {StartTime: 0}}
	sorted := []Event{{StartTime: 0}, {StartTime: 1}, {StartTime: 5}}

	m := Map{Events: unsorted}

	m.SortEvents()
	assert.Equal(t, sorted, m.Events)
}
