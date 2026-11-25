package main

import (
	"slices"
	"testing"
)

func TestCompactPropertyRows(t *testing.T) {
	// Hidden section contents were already left out. Reports and controls
	// sharing a row stay in visual order and keep their size.
	items := []panelItem{
		{handle: 4, x: 16, y: 390, height: 44},
		{handle: 1, x: 16, y: 16, height: 22},
		{handle: 2, x: 290, y: 16, height: 22},
		{handle: 3, x: 16, y: 230, height: 22},
		{handle: 5, x: 169, y: 395, height: 28},
		{handle: 6, x: 16, y: 442, height: 22},
	}
	want := []panelItem{
		{handle: 1, x: 16, y: 16, height: 22},
		{handle: 2, x: 290, y: 16, height: 22},
		{handle: 3, x: 16, y: 54, height: 22},
		{handle: 4, x: 16, y: 92, height: 44},
		{handle: 5, x: 169, y: 97, height: 28},
		{handle: 6, x: 16, y: 144, height: 22},
	}
	compactPropertyRows(items)
	if !slices.Equal(items, want) {
		t.Errorf("compact rows = %+v, want %+v", items, want)
	}
}
