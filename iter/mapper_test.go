package iter

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMappingIterator_Next(t *testing.T) {

	sl := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	sl2 := Map[int, int](Slice(sl), double)

	finished, err := ToSlice[int](sl2)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, finished, []int{2, 4, 6, 8, 10, 12, 14, 16, 18, 20})
}

func TestReducer(t *testing.T) {
	sl := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	sl2 := Map[int, int](Slice(sl), double)

	summed, err := Reduce[int](sl2, func(acc int, item int) int {
		return acc + item
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 110, summed.Or(0))
}

func double(input int) int {
	return input * 2
}
