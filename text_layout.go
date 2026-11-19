package main

import (
	"math"
	"slices"
	"strings"
	"unicode"
)

// Coordinates are in oriented PDF units with Y increasing down the page.
type layoutGlyph struct {
	run, index   int
	text         string
	bounds       searchRect
	baseline, em float64
}

type textLine struct {
	glyphs       []layoutGlyph
	bounds       searchRect
	baseline, em float64
}

type textFlow struct {
	paragraphs [][]textLine
	bounds     searchRect
	text       string
}

func mergeTextBounds(a, b searchRect) searchRect {
	if a[2] <= a[0] || a[3] <= a[1] {
		return b
	}
	return searchRect{min(a[0], b[0]), min(a[1], b[1]), max(a[2], b[2]), max(a[3], b[3])}
}

func glyphRows(glyphs []layoutGlyph) []textLine {
	slices.SortStableFunc(glyphs, func(a, b layoutGlyph) int {
		if a.baseline < b.baseline {
			return -1
		}
		if a.baseline > b.baseline {
			return 1
		}
		return 0
	})
	var rows []textLine
	for _, glyph := range glyphs {
		if n := len(rows); n > 0 {
			row := &rows[n-1]
			if math.Abs(glyph.baseline-row.baseline) <= max(.8, min(row.em, glyph.em)*.29) {
				row.glyphs = append(row.glyphs, glyph)
				continue
			}
		}
		rows = append(rows, textLine{glyphs: []layoutGlyph{glyph}, baseline: glyph.baseline, em: glyph.em})
	}
	var fragments []textLine
	for _, row := range rows {
		slices.SortStableFunc(row.glyphs, func(a, b layoutGlyph) int {
			if a.bounds[0] < b.bounds[0] {
				return -1
			}
			if a.bounds[0] > b.bounds[0] {
				return 1
			}
			return 0
		})
		fragment := textLine{baseline: row.baseline, em: row.em}
		for _, glyph := range row.glyphs {
			if len(fragment.glyphs) > 0 && glyph.bounds[0]-fragment.bounds[2] > max(8, min(fragment.em, glyph.em)*1.4) {
				fragments = append(fragments, fragment)
				fragment = textLine{baseline: glyph.baseline, em: glyph.em}
			}
			fragment.glyphs = append(fragment.glyphs, glyph)
			fragment.bounds = mergeTextBounds(fragment.bounds, glyph.bounds)
		}
		fragments = append(fragments, fragment)
	}
	return fragments
}

// Group glyphs into baseline rows and link adjacent rows on the same margin.
// Dividing rows at large gaps separates nearby columns before linking rows.
func layoutText(glyphs []layoutGlyph) textFlow {
	lines := glyphRows(glyphs)
	used := make([]bool, len(lines))
	var flow textFlow
	var paragraphs []string
	for i, line := range lines {
		if used[i] {
			continue
		}
		used[i] = true
		group := []textLine{line}
		for current := i; ; {
			next := nextTextLine(lines, used, current, line.bounds[0])
			if next < 0 {
				break
			}
			group = append(group, lines[next])
			used[next], current = true, next
		}
		var content strings.Builder
		for _, line := range group {
			flow.bounds = mergeTextBounds(flow.bounds, line.bounds)
			value := joinGlyphs(line)
			if value == "" {
				continue
			}
			if content.Len() > 0 {
				if old := content.String(); strings.HasSuffix(old, "\u00ad") {
					content.Reset()
					content.WriteString(strings.TrimSuffix(old, "\u00ad"))
				} else {
					content.WriteByte(' ')
				}
			}
			content.WriteString(value)
		}
		if content.Len() > 0 {
			flow.paragraphs = append(flow.paragraphs, group)
			paragraphs = append(paragraphs, content.String())
		}
	}
	flow.text = strings.Join(paragraphs, "\n\n")
	return flow
}

func nextTextLine(lines []textLine, used []bool, index int, margin float64) int {
	current := lines[index]
	best, score := -1, math.Inf(1)
	for i := index + 1; i < len(lines); i++ {
		candidate := lines[i]
		dy := candidate.baseline - current.baseline
		if dy > current.em*1.83 {
			break
		}
		if used[i] || dy < current.em*.55 || candidate.em < current.em*.68 || candidate.em > current.em*1.4 {
			continue
		}
		dx := math.Abs(candidate.bounds[0] - current.bounds[0])
		overlap := min(current.bounds[2], candidate.bounds[2]) - max(current.bounds[0], candidate.bounds[0])
		if overlap <= 0 || dx > current.em*1.68 || math.Abs(candidate.bounds[0]-margin) > current.em*1.8 {
			continue
		}
		if cost := dy + dx*.55; cost < score {
			best, score = i, cost
		}
	}
	return best
}

func joinGlyphs(line textLine) string {
	return writeGlyphLine(line, nil)
}

func writeGlyphLine(line textLine, span func(string, int)) string {
	var result strings.Builder
	right := math.Inf(-1)
	spaced := true
	previous := layoutGlyph{run: -1}
	for _, glyph := range line.glyphs {
		value := glyph.text
		isSpace := strings.TrimFunc(value, unicode.IsSpace) == ""
		gap := glyph.bounds[0] - right
		threshold := max(.8, line.em*.18)
		if glyph.run == previous.run && glyph.index == previous.index+1 {
			threshold = max(threshold, line.em*.32)
		}
		if !spaced && (isSpace || gap > threshold || glyph.run == previous.run && glyph.index > previous.index+1) {
			result.WriteByte(' ')
			if span != nil {
				span(" ", previous.run)
			}
			spaced = true
		}
		if !isSpace {
			result.WriteString(value)
			if span != nil {
				span(value, glyph.run)
			}
			spaced = false
		}
		right = max(right, glyph.bounds[2])
		previous = glyph
	}
	return strings.TrimSpace(result.String())
}

func (flow textFlow) glyphs() []layoutGlyph {
	var glyphs []layoutGlyph
	for _, paragraph := range flow.paragraphs {
		for _, line := range paragraph {
			glyphs = append(glyphs, line.glyphs...)
		}
	}
	return glyphs
}

func (flow textFlow) paragraphAt(hit layoutGlyph) textFlow {
	for _, group := range flow.paragraphs {
		for _, line := range group {
			for _, glyph := range line.glyphs {
				if glyph.run == hit.run && glyph.index == hit.index {
					return layoutText(textFlow{paragraphs: [][]textLine{group}}.glyphs())
				}
			}
		}
	}
	return textFlow{}
}

func textLineAt(glyphs []layoutGlyph, hit layoutGlyph) textFlow {
	for _, line := range glyphRows(glyphs) {
		for _, glyph := range line.glyphs {
			if glyph.run == hit.run && glyph.index == hit.index {
				return textFlow{paragraphs: [][]textLine{{line}}, bounds: line.bounds, text: joinGlyphs(line)}
			}
		}
	}
	return textFlow{}
}

func (flow textFlow) highlightBounds() []searchRect {
	var highlights []searchRect
	for _, group := range flow.paragraphs {
		for _, line := range group {
			highlights = append(highlights, line.bounds)
		}
	}
	return highlights
}
