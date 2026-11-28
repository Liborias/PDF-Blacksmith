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
	compactPropertyRowsWithHeaders(items,nil)
}

// Section headings already carry their own separator. Keep the surrounding
// gaps compact while preserving ordinary control spacing, alignment and sizes.
func compactPropertyRowsWithHeaders(items []panelItem, headers map[uintptr]bool) {
	slices.SortStableFunc(items, func(a, b panelItem) int { return cmp.Compare(a.y, b.y) })
	end, removed := 0, 0
	rowY, rowHeader := -1,false
	for i := range items {
		item := &items[i]
		limit := 16
		if i>0 && (headers[item.handle] || rowHeader) {
			limit=10
		}
		removed += max(0, item.y-end-limit)
		end = max(end, item.y+item.height)
		if item.y != rowY { rowY,rowHeader=item.y,false }
		rowHeader = rowHeader || headers[item.handle]
		item.y -= removed
	}
}
