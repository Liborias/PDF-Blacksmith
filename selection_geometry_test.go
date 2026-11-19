package main

import (
	"slices"
	"testing"
)

func TestDirectionalRectangle(t *testing.T) {
	area := rectangleBetween(textPoint{0.7, 0.15}, textPoint{0.2, 0.55})
	inside := searchRect{0.3, 0.25, 0.4, 0.35}
	touching := searchRect{0.7, 0.25, 0.8, 0.4}
	outside := searchRect{0.8, 0.7, 0.9, 0.9}
	if !rectangleSelects(area, inside, false) || rectangleSelects(area, touching, false) ||
		!rectangleSelects(area, touching, true) || rectangleSelects(area, outside, true) {
		t.Fatal("containment and crossing must differ at the selection border")
	}
	if rectangleTouchesSegment(area, textPoint{0, 0.95}, textPoint{0.95, 0.95}) ||
		!rectangleTouchesSegment(area, textPoint{0, 0.2}, textPoint{0.95, 0.2}) ||
		!rectangleTouchesSegment(area, textPoint{0.7, 0.55}, textPoint{0.95, 0.8}) {
		t.Fatal("path crossing must also recognize a segment or endpoint on the border")
	}
}

func TestSelectionModifiers(t *testing.T) {
	current, picked := []int{1, 3, 5}, []int{3, 7}
	for _, test := range []struct {
		mode selectionChange
		want []int
	}{
		{replaceSelection, []int{3, 7}},
		{addSelection, []int{1, 3, 5, 7}},
		{toggleSelection, []int{1, 5, 7}},
		{subtractSelection, []int{1, 5}},
	} {
		if result := combineSelection(current, picked, test.mode, func(value int) int { return value }); !slices.Equal(result, test.want) {
			t.Errorf("mode %d combined selection = %v, want %v", test.mode, result, test.want)
		}
	}
}
