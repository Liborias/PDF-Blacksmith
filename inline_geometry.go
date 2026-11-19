package main

// The width changes around the edge or centre implied by alignment. Limits
// are in page units and include the page edges and nearby text on the same row.
func inlineHorizontal(original searchRect, leftLimit, rightLimit, wanted float64, alignment int) (float64, float64) {
	anchor := original[0]
	if alignment == 1 {
		anchor = (original[0] + original[2]) / 2
	} else if alignment == 2 {
		anchor = original[2]
	}
	available := rightLimit - anchor
	switch alignment {
	case 1:
		available = 2 * min(anchor-leftLimit, rightLimit-anchor)
	case 2:
		available = anchor - leftLimit
	}
	width := min(wanted, max(1, available))
	switch alignment {
	case 1:
		return anchor - width/2, width
	case 2:
		return anchor - width, width
	}
	return anchor, width
}

func inlineNeighbourEdges(frame searchRect, glyphs []layoutGlyph, included map[[2]int]bool, width float64) (float64, float64) {
	left, right := 0., width
	span := frame[3] - frame[1]
	for _, glyph := range glyphs {
		g := glyph.bounds
		if included[[2]int{glyph.run, glyph.index}] || min(g[3], frame[3])-max(g[1], frame[1]) < min(span, g[3]-g[1])*.4 {
			continue
		}
		if g[2] <= frame[0] {
			left = max(left, min(frame[0], g[2]+1))
		}
		if g[0] >= frame[2] {
			right = min(right, max(frame[2], g[0]-1))
		}
	}
	return left, right
}
