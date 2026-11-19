package main

import (
	"math"
	"testing"
)

func TestSearchDisplayRectClipsAndRotatesWithPageBox(t *testing.T) {
	box := searchRect{100, 200, 300, 600}
	hit := searchRect{90, 400, 150, 500}
	for angle, want := range map[int]searchRect{
		0:   {0, 0.25, 0.25, 0.5},
		90:  {0.5, 0, 0.75, 0.25},
		180: {0.75, 0.5, 1, 0.75},
		270: {0.25, 0.75, 0.5, 1},
	} {
		got, ok := searchDisplayRect(hit, box, angle)
		if !ok {
			t.Fatalf("no rectangle at %d°", angle)
		}
		for index, value := range got {
			if math.Abs(value-want[index]) > 1e-9 {
				t.Fatalf("at %d°: got %v, want %v", angle, got, want)
			}
		}
	}
	if _, ok := searchDisplayRect(searchRect{1, 2, 3, 4}, box, 90); ok {
		t.Fatal("text fully outside the crop should have no visible highlight")
	}
}
