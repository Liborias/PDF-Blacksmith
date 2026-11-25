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

func TestReadingZoomPresets(t *testing.T) {
	page := pageSize{360, 720}
	for _, tc := range []struct {
		mode readingZoomMode
		dpi  int
		want float64
	}{
		{readingZoomFitPage, 96, 460.0 / 720},
		{readingZoomFitWidth, 96, 560.0 / 360},
		{readingZoomActual, 96, 96.0 / 72},
		{readingZoomActual, 144, 2},
		{readingZoomActual, 192, 192.0 / 72},
		{readingZoomManual, 96, 3},
	} {
		if got := readingZoomScale(tc.mode, 3, 600, 500, 20, page, tc.dpi); math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("mode %d at DPI %d yields %.4f, want %.4f", tc.mode, tc.dpi, got, tc.want)
		}
	}
}

func TestReadingZoomUsesPercentLimitsWithoutJumpingFromAnExtremeFit(t *testing.T) {
	const dpi = 144
	actual := pixelsPerPDFPoint(dpi)
	if got := stepReadingZoom(actual, -50, dpi); got != actual*0.30 {
		t.Fatalf("minimum manual scale = %g", got)
	}
	if got := stepReadingZoom(actual, 50, dpi); got != actual*9 {
		t.Fatalf("maximum manual scale = %g", got)
	}
	smallFit := actual * 0.20
	if got := stepReadingZoom(smallFit, -1, dpi); got != smallFit {
		t.Fatalf("zooming out from a very small fit jumped to %g", got)
	}
	if got := stepReadingZoom(smallFit, 1, dpi); got <= smallFit || got >= actual*0.30 {
		t.Fatalf("zooming in from a very small fit jumped to %g", got)
	}
}

func TestThumbnailZoomAccumulatesSmallWheelDeltasAndClamps(t *testing.T) {
	scale, remainder := 32, 0
	for i := 0; i < 12; i++ {
		scale, remainder = stepThumbnailZoom(scale, 10, remainder)
	}
	if scale != 37 || remainder != 0 {
		t.Fatalf("one wheel detent in twelve pulses yields %d%% with %d remainder", scale, remainder)
	}
	scale, remainder = stepThumbnailZoom(59, 50, 0)
	if scale != 60 || remainder != 0 {
		t.Fatalf("zoom-in limit retained outward input: %d%% with %d remainder", scale, remainder)
	}
	scale, remainder = stepThumbnailZoom(11, -50, 0)
	if scale != 10 || remainder != 0 {
		t.Fatalf("zoom-out limit retained outward input: %d%% with %d remainder", scale, remainder)
	}
}

func TestReadingPageTurnRequiresReachingTheEdgeInScrollDirection(t *testing.T) {
	for _, tc := range []struct {
		offset, content, viewport, distance float64
		direction                           int
	}{
		{0, 1400, 700, 1200, 0}, // Even a large move from the top scrolls this page first.
		{695, 1400, 700, 105, 0},
		{700, 1400, 700, 105, 1},
		{700, 1400, 700, -105, 0},
		{5, 1400, 700, -105, 0},
		{0, 1400, 700, -105, -1},
		{-30, 600, 700, 105, 1},
		{-30, 600, 700, -105, -1},
		{-30, 600, 700, 0, 0},
	} {
		if got := readingPageTurn(tc.offset, tc.content, tc.viewport, tc.distance); got != tc.direction {
			t.Errorf("page turn for %+v = %d", tc, got)
		}
	}
}

func TestPageWheelLetsOneTurnThroughAndScrollsWithinTheNextPage(t *testing.T) {
	var gesture pageWheelGesture
	if !gesture.allow(-360, 10, true) {
		t.Fatal("first edge event should turn the page")
	}
	for tick := uint32(110); tick <= 810; tick += 100 {
		if gesture.allow(-120, tick, true) {
			t.Fatalf("a second turn escaped the same burst at %d ms", tick)
		}
	}
	if !gesture.allow(-35, 890, false) || gesture.allow(-120, 950, true) {
		t.Fatal("the same gesture should scroll within a tall page without turning again")
	}
	if !gesture.allow(-120, 1250, true) {
		t.Fatal("a pause should let the next turn through")
	}
	if !gesture.allow(120, 1260, true) {
		t.Fatal("reversing direction should start a fresh gesture")
	}
}

func TestPageWheelCountsOnlyAdjacentPartialDeltasAtTheEdge(t *testing.T) {
	var gesture pageWheelGesture
	gesture.allow(-115, 5, false)
	if gesture.allow(-115, 15, true) || !gesture.allow(-5, 25, true) {
		t.Fatal("distance traveled inside the page must not count toward turning it")
	}
	gesture = pageWheelGesture{}
	if gesture.allow(70, 50, true) || gesture.allow(60, 450, true) ||
		gesture.allow(-60, 470, true) || !gesture.allow(-60, 490, true) {
		t.Fatal("partial deltas must restart after a pause or direction reversal")
	}
	gesture = pageWheelGesture{}
	start := ^uint32(0) - 10
	if !gesture.allow(120, start, true) || gesture.allow(120, 15, true) {
		t.Fatal("a Win32 tick counter wrap must keep adjacent events in one gesture")
	}
}
