package main

const (
	idLayoutAids = 233 + iota
	idToggleLayoutGrid
)

type layoutAidsPanel struct {
	grid      uintptr
	temporary bool
	revision  uint64
}

func (a *app) buildLayoutAidsPanel() {
	a.beginPropertySection(uiLayoutGrid, 15, uiLayoutGridHelp)
	a.layoutAids.grid = a.uiButton(uiShowLayoutGrid, idToggleLayoutGrid, 16, 43, 298, bsAutoCheck)
}

func (a *app) toggleLayoutGrid() {
	if a.busy || !a.view.ready() {
		return
	}
	p := &a.layoutAids
	a.setChecked(p.grid, !a.checked(p.grid))
	if a.mode != idLayoutAids {
		p.temporary = true
		p.revision++
		a.layout()
	}
	a.view.invalidateCanvas()
	if a.comparisonVisible() {
		a.comparison.view.invalidateCanvas()
	}
}

func (a *app) effectiveRightDock() dockVisibility {
	dock := a.rightDock
	if a.layoutAids.temporary {
		dock.revealPrimary()
	}
	return dock
}

func (a *app) layoutAidsOutsideClick(msg *winMessage) uint64 {
	if !a.layoutAids.temporary || (msg.message != 0x0202 && msg.message != 0x0205) {
		return 0 // Only completed left or right mouse clicks can dismiss the context.
	}
	child, _, _ := user32.NewProc("IsChild").Call(a.sidebar, msg.window)
	if child != 0 || msg.window == a.sidebar || msg.window == a.sidebarTitle {
		return 0
	}
	return a.layoutAids.revision
}

func (a *app) dismissLayoutAids() bool {
	if !a.layoutAids.temporary {
		return false
	}
	a.layoutAids.temporary = false
	a.layout()
	return true
}

func (v *documentView) paintLayoutGrid(dc uintptr, page rect, size pageSize) {
	visible := rect{
		max(page.left, v.viewport.left), max(page.top, v.viewport.top),
		min(page.right, v.viewport.right), min(page.bottom, v.viewport.bottom),
	}
	if visible.left >= visible.right || visible.top >= visible.bottom {
		return
	}
	mask, _, _ := createSurfaceDC.Call(dc)
	if mask == 0 {
		return
	}
	defer deleteSurfaceDC.Call(mask)
	bitmap, _, _ := createSurfaceBitmap.Call(dc, 1, 1)
	if bitmap == 0 {
		return
	}
	defer gdiDelete.Call(bitmap)
	old, _, _ := gdiSelect.Call(mask, bitmap)
	if old == 0 || old == ^uintptr(0) {
		return
	}
	defer gdiSelect.Call(mask, old)
	fillColor(mask, rect{right: 1, bottom: 1}, paletteColor{83, 127, 153})

	paint := func(x, y, width, height int32, major bool) {
		opacity := uintptr(62)
		if major {
			opacity = 119
		}
		alphaBlend.Call(dc, uintptr(x), uintptr(y), uintptr(width), uintptr(height), mask, 0, 0, 1, 1, opacity<<16)
	}
	gap := float64(v.app.scale(5))
	visitLayoutGridLines(page.left, page.right, visible.left, visible.right, size.width, gap, func(x int32, major bool) {
		paint(x, visible.top, 1, visible.bottom-visible.top, major)
	})
	visitLayoutGridLines(page.top, page.bottom, visible.top, visible.bottom, size.height, gap, func(y int32, major bool) {
		paint(visible.left, y, visible.right-visible.left, 1, major)
	})
}
