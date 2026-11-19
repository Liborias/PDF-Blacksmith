package main

import (
	"math"
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

func TestParagraphJoinsWordsAndStopsAtWhitespace(t *testing.T) {
	word := func(index int, value string, x, y, width float64) layoutGlyph {
		return layoutGlyph{run: index, text: value, bounds: searchRect{x, y - 10, x + width, y + 2}, baseline: y, em: 12}
	}
	glyphs := []layoutGlyph{
		word(0, "jiný sloupec", 290, 110, 90),
		word(1, "První", 40, 110, 28), word(2, "řádek", 73, 110, 30),
		word(3, "pokračuje", 40, 126, 50), word(4, "tady.", 95, 126, 24),
		word(5, "Druhý odstavec.", 40, 155, 110),
		word(6, "navazuje", 290, 126, 65),
	}
	result := layoutText(glyphs).paragraphAt(glyphs[1])
	if result.text != "První řádek pokračuje tady." {
		t.Fatalf("paragraph = %q", result.text)
	}
	if all := layoutText(glyphs); len(all.paragraphs) != 3 {
		t.Fatalf("expected three separated blocks, have %q", all.text)
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
