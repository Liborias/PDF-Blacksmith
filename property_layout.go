package main

import (
	"cmp"
	"slices"
)

type panelItem struct {
	handle              uintptr
	x, y, width, height int
	section             int
}

// Compact only gaps between rows. Controls on the same row keep their alignment.
func compactPropertyRows(items []panelItem) {
	slices.SortStableFunc(items, func(a, b panelItem) int { return cmp.Compare(a.y, b.y) })
	end, removed := 0, 0
	for i := range items {
		item := &items[i]
		removed += max(0, item.y-end-16)
		end = max(end, item.y+item.height)
		item.y -= removed
	}
}
