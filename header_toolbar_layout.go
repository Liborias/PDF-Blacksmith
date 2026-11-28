package main

import "fmt"

type headerToolbarPlan struct {
	iconWidth, gap, search, zoom, used int
	hidden map[int]bool
}

// available excludes the measured text menus and outer spacing. Charge a
// trailing gap for each visible item to keep the arithmetic conservative.
func planHeaderToolbar(available int, icons, optional, emergency []int,
	normal, normalGap, compact, compactGap, searchSmall, searchMin, searchMax, zoomSize int) headerToolbarPlan {
	p := headerToolbarPlan{iconWidth: normal, gap: normalGap, hidden: make(map[int]bool)}
	present := make(map[int]bool, len(icons))
	for _, id := range icons {
		present[id] = true
	}
	count := len(present)
	iconCost := func() int { return count * (p.iconWidth + p.gap) }
	desired := func() int { return iconCost() + searchMax + zoomSize + 2*p.gap }
	if desired() > available {
		p.iconWidth, p.gap = compact, compactGap
	}
	hide := func(id int) {
		if present[id] && !p.hidden[id] {
			p.hidden[id] = true
			count--
		}
	}
	// Hide optional duplicates in the approved order before shrinking search.
	for _, id := range optional {
		if desired() <= available {
			break
		}
		hide(id)
	}
	// Only hide further commands if even the icon-only bar cannot fit.
	for _, id := range emergency {
		if iconCost() <= available {
			break
		}
		hide(id)
	}
	free := max(0, available-iconCost())
	p.search, p.zoom = headerOptionalWidths(free, p.gap, searchMin, searchMax, zoomSize)
	if p.search == 0 && free >= searchSmall+p.gap {
		// Prefer a compact search input to the zoom dropdown at this width.
		p.search, p.zoom = min(searchMax, free-p.gap), 0
		if free-p.search-p.gap >= zoomSize+p.gap {
			p.zoom = zoomSize
		}
	}
	p.used = iconCost()
	if p.search > 0 {
		p.used += p.search + p.gap
	}
	if p.zoom > 0 {
		p.used += p.zoom + p.gap
	}
	return p
}

// The page group is supplementary. Keep all existing visible commands and
// the search/zoom fallback policy before accepting a narrower toolbar plan.
func headerPageNavigationFits(base, candidate headerToolbarPlan, searchMin int) bool {
	for id := range candidate.hidden {
		if !base.hidden[id] {
			return false
		}
	}
	return (base.search == 0 || candidate.search >= searchMin) &&
		(base.zoom == 0 || candidate.zoom > 0)
}

func headerPageNavigationState(ready, busy bool, page, total int) (string, bool, bool) {
	if !ready || total <= 0 || page < 0 || page >= total {
		return "— / —", false, false
	}
	return fmt.Sprintf("%d / %d", page+1, total), !busy && page > 0, !busy && page+1 < total
}
