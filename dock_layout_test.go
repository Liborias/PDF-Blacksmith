package main

import "testing"

func TestFitDockWidthsKeepsPanelsAndDocumentUsable(t *testing.T) {
	for _, dpi := range []int{96, 144, 192} {
		scale := func(px int) int { return px * dpi / 96 }
		minimum := dockWidths{scale(168), scale(350), scale(70), scale(145)}
		preferred := dockWidths{scale(195), scale(454), scale(220), scale(330)}
		last := minimum
		for _, width := range []int{980, 1100, 1250, 1600} {
			space := scale(width) - 2*scale(28)
			got := fitDockWidths(space, scale(176), scale(265), minimum, preferred)
			document := space
			for panel, size := range got {
				document -= size
				if size < last[panel] || size > preferred[panel] {
					t.Errorf("%d DPI, window %d: panel %d got %d with bounds %d to %d", dpi, width, panel, size, last[panel], preferred[panel])
				}
			}
			if document < scale(176) {
				t.Errorf("%d DPI, window %d: document is only %d px wide", dpi, width, document)
			}
			last = got
		}
		if last != preferred {
			t.Errorf("%d DPI: the widest window got %v, want %v", dpi, last, preferred)
		}
	}
}

func TestFitDockWidthsLeavesHiddenPanelsOutOfTheLayout(t *testing.T) {
	const space = 980 - 2*28
	minimum := dockWidths{dockProperties: 350, dockBookmarks: 145}
	preferred := dockWidths{dockProperties: 454, dockBookmarks: 330}
	got := fitDockWidths(space, 176, 265, minimum, preferred)
	if got[dockTools] != 0 || got[dockThumbnails] != 0 {
		t.Fatalf("the collapsed dock still takes space: %v", got)
	}
	if remaining := space - got[dockProperties] - got[dockBookmarks]; remaining != 265 {
		t.Errorf("document width after panel growth = %d, want 265", remaining)
	}

	minimum = dockWidths{168, 350, 70, 145}
	preferred = dockWidths{195, 454, 220, 330}
	got = fitDockWidths(600, 176, 265, minimum, preferred)
	if got[dockThumbnails] != 0 || got[dockBookmarks] != 0 || got[dockTools]+got[dockProperties] != 600-176 {
		t.Errorf("limited space should preserve the document and main panels first: %v", got)
	}
}

func TestDockCollapsePreservesIndividualPanels(t *testing.T) {
	dock := dockVisibility{primaryCollapsed: true}
	dock.collapsed = true
	if dock.showsPrimary() || dock.showsAuxiliary(true) {
		t.Fatal("a collapsed dock still shows a panel")
	}
	dock.collapsed = false
	if dock.showsPrimary() || dock.showsAuxiliary(false) {
		t.Fatal("reopening a dock also opened an individually closed or unavailable panel")
	}
	if !dock.showsAuxiliary(true) {
		t.Fatal("the auxiliary panel failed to return when it became available")
	}

	for _, auxiliaryCollapsed := range []bool{false, true} {
		dock = dockVisibility{collapsed: true, auxiliaryCollapsed: auxiliaryCollapsed, primaryCollapsed: true}
		dock.revealPrimary()
		if !dock.showsPrimary() || dock.showsAuxiliary(true) != !auxiliaryCollapsed {
			t.Error("revealing the primary panel should preserve the auxiliary panel's own setting")
		}
	}
}
