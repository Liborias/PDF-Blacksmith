package main

import (
	"slices"
	"testing"
)

func TestOverviewWrapsAndHitsOnlyRealCards(t *testing.T) {
	for _, tc := range []struct {
		width, cardWidth, columns int
	}{{630, 173, 3}, {420, 173, 2}, {420, 115, 3}, {140, 173, 1}} {
		grid := layoutOverview(7, tc.width, tc.cardWidth, 12, 28)
		if grid.columns != tc.columns {
			t.Fatalf("layout %+v got %d columns", tc, grid.columns)
		}
		for index := 0; index < grid.count; index++ {
			x, y := grid.card(index)
			if hit := grid.at(x+grid.cardWidth/2, y+grid.cardHeight/2); hit != index {
				t.Fatalf("card %d was hit as %d", index, hit)
			}
			if hit := grid.at(x+grid.cardWidth, y); hit != -1 {
				t.Fatalf("gap after card %d hit as %d", index, hit)
			}
		}
		x, y := grid.card(grid.count)
		if hit := grid.at(x, y); hit != -1 {
			t.Fatalf("unused slot was hit as %d", hit)
		}
	}
}

func TestOverviewRestrictsTheVisibleRangeForLargePDFs(t *testing.T) {
	grid := layoutOverview(18000, 1200, 145, 12, 28)
	_, y := grid.card(9000)
	first, last := grid.visible(y, 640)
	if first > 9000 || last <= 9000 || last-first > 45 {
		t.Fatalf("unexpected visible range %d to %d", first, last)
	}
	_, last = grid.visible(grid.height-500, 500)
	if last != grid.count {
		t.Fatalf("last row was clipped at %d of %d", last, grid.count)
	}
}

func TestOverviewSelectionKeepsSortedPositionsAndEarlierSelection(t *testing.T) {
	initial := []int{1, 5}
	added := overviewSelection(initial, 1, 3, true, false)
	if !slices.Equal(added, []int{1, 3, 5}) || !slices.Equal(initial, []int{1, 5}) {
		t.Fatalf("Ctrl+click gave %v, previous selection %v", added, initial)
	}
	if got := overviewSelection(added, 3, 1, true, false); !slices.Equal(got, []int{3, 5}) {
		t.Fatalf("Ctrl+click did not remove %v", got)
	}
	if got := overviewSelection(added, 5, 2, false, true); !slices.Equal(got, []int{2, 3, 4, 5}) {
		t.Fatalf("Shift+click gave %v", got)
	}
	if got := overviewSelection([]int{1}, 1, 1, true, false); len(got) != 0 {
		t.Fatalf("last selected card could not be deselected: %v", got)
	}
}
