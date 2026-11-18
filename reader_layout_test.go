package main

import (
	"math"
	"testing"
)

func TestCursorStaysAtSamePointWhenZooming(t *testing.T) {
	sizes := []pageSize{{595, 842}, {842, 595}, {595, 842}}
	before := layoutPages(sizes, 1, true, 0.9, 18)
	x, y := 350.0, 990.0
	point, ok := before.anchor(x, y)
	if !ok || point.index != 1 {
		t.Fatalf("unexpected anchor: %+v", point)
	}
	after := layoutPages(sizes, 1, true, 1.7, 18)
	panX, panY, ok := after.restore(point, 180, 235)
	active, _ := after.anchor(panX+180, panY+235)
	if !ok || active.index != point.index || math.Abs(active.x-point.x) > 1e-8 || math.Abs(active.y-point.y) > 1e-8 {
		t.Fatalf("zoom moved cursor from %+v to %+v", point, active)
	}
}

func TestContinuousLayoutShowsAdjacentPages(t *testing.T) {
	sizes := []pageSize{{595, 842}, {595, 842}, {842, 595}}
	spread := layoutPages(sizes, 0, true, 1, 18)
	if pages := spread.visible(750, 400); len(pages) != 2 || pages[0].index != 0 || pages[1].index != 1 {
		t.Fatalf("unexpected visible pages: %+v", pages)
	}
	page := layoutPages(sizes, 2, false, 1, 18)
	if len(page.sheets) != 1 || page.sheets[0].index != 2 || page.width != 842+36 {
		t.Fatalf("unexpected single page layout: %+v", page)
	}
}

func TestReadingOffsetAllowsCursorAnchorWhilePageFits(t *testing.T) {
	if got := clampReadingOffset(-60, 600, 800); got != -60 {
		t.Fatalf("small page lost anchor: %v", got)
	}
	if got := clampReadingOffset(325, 1000, 800); got != 200 {
		t.Fatalf("large page scrolled outside its edges: %v", got)
	}
}
