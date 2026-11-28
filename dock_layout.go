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

type auxiliaryPanel uint8

const (
	auxiliaryNone auxiliaryPanel = iota
	auxiliaryBookmarks
	auxiliaryResults
	auxiliaryLayers
)

type rightAuxiliary struct {
	panel        auxiliaryPanel
	beforeSearch *rightDockSnapshot
}

type rightDockSnapshot struct {
	panel auxiliaryPanel
	dock  dockVisibility
}

func (aux *rightAuxiliary) syncSearch(active bool, dock *dockVisibility) {
	if active && aux.beforeSearch == nil {
		aux.beforeSearch = &rightDockSnapshot{aux.panel, *dock}
		aux.panel = auxiliaryResults
		dock.revealPrimary()
	}
	if !active && aux.beforeSearch != nil {
		aux.panel, *dock = aux.beforeSearch.panel, aux.beforeSearch.dock
		aux.beforeSearch = nil
	}
}

func (aux *rightAuxiliary) togglePanel(panel auxiliaryPanel) {
	if aux.panel == panel {
		aux.panel = auxiliaryNone
	} else {
		aux.panel = panel
	}
	if aux.beforeSearch != nil {
		aux.beforeSearch.panel = aux.panel
	}
}

func (aux *rightAuxiliary) setGroupCollapsed(dock *dockVisibility, collapsed bool) {
	dock.collapsed = collapsed
	if aux.beforeSearch != nil {
		aux.beforeSearch.dock.collapsed = collapsed
	}
}

func (aux *rightAuxiliary) setPropertiesCollapsed(dock *dockVisibility, collapsed bool) {
	dock.primaryCollapsed = collapsed
	if aux.beforeSearch != nil {
		aux.beforeSearch.dock.primaryCollapsed = collapsed
	}
}

const (
	dockTools = iota
	dockProperties
	dockThumbnails
	dockRightAuxiliary
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

// The temporary wide-document view is an overlay, not a change to the user's
// saved group, primary or auxiliary visibility. Search snapshots remain intact.
func temporaryDockVisibility(dock dockVisibility, collapsed bool) dockVisibility {
	if collapsed {
		dock.collapsed = true
	}
	return dock
}


// Keep ears outside the document rectangle, after all visible dock panels.
// The two gutters belong to the editor background, not to the panels.
type dockEdges struct {
	leftAuxStart, leftAuxEnd int
	documentStart, documentEnd int
	rightAuxStart, rightAuxEnd int
}

func dockDocumentEdges(width, gutter int, sizes dockWidths) dockEdges {
	left := sizes[dockTools]
	leftEnd := left + sizes[dockThumbnails]
	right := width - sizes[dockProperties]
	rightStart := right - sizes[dockRightAuxiliary]
	return dockEdges{left, leftEnd, leftEnd + gutter, rightStart - gutter, rightStart, right}
}
