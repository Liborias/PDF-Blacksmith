package main

import "testing"

func TestInlineGrowthKeepsTheAlignmentAnchorAndStaysBetweenNeighbours(t *testing.T) {
	frame := searchRect{60, 20, 90, 30}
	glyphs := []layoutGlyph{
		{run: 1, bounds: searchRect{26, 20, 40, 30}},
		{run: 2, bounds: searchRect{99, 20, 108, 30}},
		{run: 3, bounds: searchRect{90, 52, 93, 58}},
		{run: 4, bounds: searchRect{92, 20, 98, 30}},
	}
	left, right := inlineNeighbourEdges(frame, glyphs, map[[2]int]bool{{4, 0}: true}, 210)
	if left != 41 || right != 98 {
		t.Fatalf("limits %g, %g", left, right)
	}
	for _, c := range []struct {
		align                int
		natural, left, width float64
	}{
		{0, 200, 60, 38}, {1, 200, 52, 46}, {2, 200, 41, 49}, {3, 200, 60, 38},
		{0, 15, 60, 15}, {1, 10, 70, 10}, {2, 10, 80, 10},
	} {
		gotLeft, gotWidth := inlineHorizontal(frame, left, right, c.natural, c.align)
		if gotLeft != c.left || gotWidth != c.width {
			t.Errorf("alignment %d width %g: got (%g,%g), want (%g,%g)", c.align, c.natural, gotLeft, gotWidth, c.left, c.width)
		}
	}
}
