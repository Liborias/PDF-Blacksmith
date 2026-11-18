package main

import (
	"math"
	"sort"
)

type pageSize struct{ width, height float64 }

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
	first, last := current, current+1
	if continuous {
		first, last = 0, len(sizes)
	}
	width := 0.0
	for _, page := range sizes[first:last] {
		width = max(width, page.width*zoom)
	}
	layout := readingLayout{width: width + 2*gap, sheets: make([]sheet, 0, last-first)}
	top := gap
	for index := first; index < last; index++ {
		w, h := sizes[index].width*zoom, sizes[index].height*zoom
		layout.sheets = append(layout.sheets, sheet{index, (layout.width - w) / 2, top, w, h})
		top += h + gap
	}
	layout.height = top
	return layout
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
