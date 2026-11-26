package main

import (
	"math"
	"sort"
)

type pageSize struct{ width, height float64 }

type readingZoomMode int

const (
	readingZoomManual readingZoomMode = iota
	readingZoomFitPage
	readingZoomFitWidth
	readingZoomActual
)

func pixelsPerPDFPoint(dpi int) float64 {
	return float64(dpi) / 72
}

func readingZoomScale(mode readingZoomMode, current, width, height, margin float64, page pageSize, dpi int) float64 {
	switch {
	case mode == readingZoomActual:
		return pixelsPerPDFPoint(dpi)
	case mode == readingZoomManual && current > 0:
		return current
	}
	fitWidth := max(0.02, (width-2*margin)/page.width)
	if mode == readingZoomFitWidth {
		return fitWidth
	}
	return min(fitWidth, max(0.02, (height-2*margin)/page.height))
}

func stepReadingZoom(current, steps float64, dpi int) float64 {
	actual := pixelsPerPDFPoint(dpi)
	// Fit modes may need to go beyond the manual range to show an entire page.
	return min(max(current*math.Pow(1.2, steps), min(current, actual*0.30)), max(current, actual*9))
}

func stepThumbnailZoom(scale, delta, remainder int) (int, int) {
	const minimum, maximum = 10, 60
	remainder += 5 * delta
	scale = min(maximum, max(minimum, scale+remainder/wheelDetent))
	remainder %= wheelDetent
	if (scale == minimum && remainder < 0) || (scale == maximum && remainder > 0) {
		remainder = 0
	}
	return scale, remainder
}

type sheet struct {
	index                    int
	left, top, width, height float64
}

type readingLayout struct {
	sheets        []sheet
	width, height float64
}

type readingAnchor struct {
	index int
	x, y  float64 // Position in the page, as a fraction of its size.
}

func layoutPages(sizes []pageSize, current int, continuous bool, zoom, gap float64) readingLayout {
	if len(sizes) == 0 || current < 0 || current >= len(sizes) {
		return readingLayout{}
	}
	if !continuous {
		return layoutPage(sizes[current], current, zoom, gap)
	}
	width := 0.0
	for _, page := range sizes {
		width = max(width, page.width*zoom)
	}
	layout := readingLayout{width: width + 2*gap, sheets: make([]sheet, 0, len(sizes))}
	top := gap
	for index, size := range sizes {
		w, h := size.width*zoom, size.height*zoom
		layout.sheets = append(layout.sheets, sheet{index, (layout.width - w) / 2, top, w, h})
		top += h + gap
	}
	layout.height = top
	return layout
}

func layoutPage(size pageSize, index int, zoom, gap float64) readingLayout {
	w, h := size.width*zoom, size.height*zoom
	return readingLayout{[]sheet{{index, gap, gap, w, h}}, w + 2*gap, h + 2*gap}
}

func (layout readingLayout) position(index int) (sheet, bool) {
	if len(layout.sheets) == 0 {
		return sheet{}, false
	}
	offset := index - layout.sheets[0].index
	if offset < 0 || offset >= len(layout.sheets) {
		return sheet{}, false
	}
	return layout.sheets[offset], true
}

func (layout readingLayout) visible(top, height float64) []sheet {
	first := sort.Search(len(layout.sheets), func(i int) bool {
		s := layout.sheets[i]
		return s.top+s.height >= top
	})
	last := first
	for last < len(layout.sheets) && layout.sheets[last].top <= top+height {
		last++
	}
	return layout.sheets[first:last]
}

func (layout readingLayout) anchor(x, y float64) (readingAnchor, bool) {
	if len(layout.sheets) == 0 {
		return readingAnchor{}, false
	}
	index := sort.Search(len(layout.sheets), func(i int) bool {
		s := layout.sheets[i]
		return s.top+s.height >= y
	})
	index = min(index, len(layout.sheets)-1)
	s := layout.sheets[index]
	if index > 0 && y < s.top {
		previous := layout.sheets[index-1]
		if y-previous.top-previous.height < s.top-y {
			s = previous
		}
	}
	return readingAnchor{s.index, (x - s.left) / s.width, (y - s.top) / s.height}, true
}

func (layout readingLayout) restore(anchor readingAnchor, cursorX, cursorY float64) (float64, float64, bool) {
	s, ok := layout.position(anchor.index)
	if !ok {
		return 0, 0, false
	}
	return s.left + anchor.x*s.width - cursorX, s.top + anchor.y*s.height - cursorY, true
}

func clampReadingOffset(value, content, viewport float64) float64 {
	edge := content - viewport
	return math.Max(math.Min(0, edge), math.Min(math.Max(0, edge), value))
}

func readingPageTurn(offset, content, viewport, distance float64) int {
	edge := content - viewport
	switch {
	case distance > 0 && (edge <= 0 || offset >= edge-0.5):
		return 1
	case distance < 0 && (edge <= 0 || offset <= 0.5):
		return -1
	default:
		return 0
	}
}

const (
	wheelDetent     = 120
	pageWheelIdleMS = 300
)

type pageWheelGesture struct {
	lastTick  uint32
	direction int
	edgeDelta int
	turned    bool
}

// Preserve one page turn per same-direction burst of Win32 wheel messages.
// Only deltas received at the edge count toward a turn.
func (gesture *pageWheelGesture) allow(delta int, tick uint32, atEdge bool) bool {
	if delta == 0 {
		return false
	}
	direction := 1
	if delta < 0 {
		direction, delta = -1, -delta
	}
	if gesture.direction != direction || tick-gesture.lastTick >= pageWheelIdleMS {
		*gesture = pageWheelGesture{direction: direction}
	}
	gesture.lastTick = tick
	if !atEdge {
		gesture.edgeDelta = 0
		return true
	}
	if gesture.turned {
		return false
	}
	gesture.edgeDelta += delta
	if gesture.edgeDelta < wheelDetent {
		return false
	}
	gesture.turned = true
	return true
}
