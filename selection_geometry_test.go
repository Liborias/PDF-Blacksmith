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

func TestLassoIntersectsContentAtItsEdge(t *testing.T) {
	lasso := []textPoint{{0.1, 0.1}, {0.8, 0.1}, {0.1, 0.8}}
	for _, test := range []struct {
		name string
		box  searchRect
		hit  bool
	}{
		{"content inside", searchRect{0.2, 0.2, 0.24, 0.24}, true},
		{"content with center outside", searchRect{0.35, 0.45, 0.57, 0.65}, true},
		{"only the lasso bounding box overlaps", searchRect{0.6, 0.6, 0.7, 0.7}, false},
		{"lasso entirely inside", searchRect{0, 0, 1, 1}, true},
	} {
		if got := polygonTouchesRect(lasso, test.box); got != test.hit {
			t.Errorf("%s: got %v, want %v", test.name, got, test.hit)
		}
	}
}

func TestLassoContainsAndCrossingStaysOpen(t *testing.T) {
	lasso := []textPoint{{0, 0}, {8, 0}, {8, 8}, {5, 8}, {5, 3}, {3, 3}, {3, 8}, {0, 8}}
	if polygonContainsRect(lasso, searchRect{1, 2, 7, 4}) {
		t.Fatal("a rectangle intersecting the notch must not be contained")
	}
	if !polygonContainsRect(lasso, searchRect{1, 4, 2, 6}) {
		t.Fatal("lasso failed to contain a rectangle on one side of the notch")
	}
	line := []textPoint{{0, 0}, {8, 0}, {8, 8}}
	if !polylineTouchesRect(line, searchRect{7, 3, 9, 4}) ||
		polylineTouchesRect(line, searchRect{3, 3, 5, 5}) {
		t.Fatal("crossing must test only the drawn segments")
	}
}
