package main

import (
	"math"
	"slices"
	"testing"
)

func TestTextRegionUsesDisplayedRotation(t *testing.T) {
	box := searchRect{50, 70, 280, 320}
	run := searchRect{100, 130, 140, 170}
	for _, rotation := range []int{0, 90, 180, 270} {
		displayed, ok := searchDisplayRect(run, box, rotation)
		if !ok {
			t.Fatalf("missing text in rotation %d", rotation)
		}
		center := textPoint{(displayed[0] + displayed[2]) / 2, (displayed[1] + displayed[3]) / 2}
		point := textPointToPDF(center, box, rotation)
		if math.Abs(point[0]-120) > 1e-9 || math.Abs(point[1]-150) > 1e-9 {
			t.Fatalf("rotation %d produced %v", rotation, point)
		}
	}
}

func TestParagraphStaysInItsColumn(t *testing.T) {
	bounds := []searchRect{
		{.51, .1, .85, .13},
		{.10, .15, .43, .18},
		{.10, .10, .45, .13},
		{.50, .15, .92, .18},
		{.10, .32, .42, .35},
	}
	if got := textParagraph(bounds, 1); !slices.Equal(got, []int{2, 1}) {
		t.Fatalf("paragraph = %v", got)
	}
}

func TestThumbnailScopeUsesPagePositions(t *testing.T) {
	cases := []struct {
		positions  []int
		contiguous bool
	}{
		{nil, false},
		{[]int{3}, false},
		{[]int{0, 2, 3}, false},
		{[]int{1, 2, 3}, true},
	}
	for _, test := range cases {
		numbers := pageNumbers(test.positions)
		for i, number := range numbers {
			if number != test.positions[i]+1 {
				t.Fatalf("page numbers = %v for %v", numbers, test.positions)
			}
		}
		if got := contiguousPages(numbers); got != test.contiguous {
			t.Fatalf("contiguous = %t for %v", got, numbers)
		}
	}
}
