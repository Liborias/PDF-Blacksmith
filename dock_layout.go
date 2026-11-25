package main

// The dock can be hidden without changing either panel's own visibility.
type dockVisibility struct {
	collapsed, primaryCollapsed, auxiliaryCollapsed bool
}

func (dock dockVisibility) showsPrimary() bool {
	return !dock.collapsed && !dock.primaryCollapsed
}

func (dock dockVisibility) showsAuxiliary(available bool) bool {
	return !dock.collapsed && !dock.auxiliaryCollapsed && available
}

func (dock *dockVisibility) revealPrimary() {
	dock.collapsed, dock.primaryCollapsed = false, false
}

const (
	dockTools = iota
	dockProperties
	dockThumbnails
	dockBookmarks
	dockPanelCount
)

type dockWidths [dockPanelCount]int

func fitDockWidths(space, minimumDocument, preferredDocument int, minimum, preferred dockWidths) dockWidths {
	budget := max(0, space-minimumDocument)
	var widths dockWidths
	// Keep the main tools usable, then reserve the auxiliary panels' compact sizes.
	budget -= widths.grow(dockWidths{dockTools: minimum[dockTools], dockProperties: minimum[dockProperties]}, budget)
	budget -= widths.grow(minimum, budget)
	// Extra width can follow the content only after the document has enough room.
	gutter := max(0, preferredDocument-minimumDocument)
	widths.grow(preferred, max(0, budget-gutter))
	return widths
}

func (widths *dockWidths) grow(targets dockWidths, budget int) int {
	var missing dockWidths
	remaining := 0
	for panel, target := range targets {
		missing[panel] = max(0, target-widths[panel])
		remaining += missing[panel]
	}
	used := min(budget, remaining)
	budget = used
	for panel, need := range missing {
		if need == 0 {
			continue
		}
		added := budget * need / remaining
		widths[panel] += added
		budget -= added
		remaining -= need
	}
	return used
}
