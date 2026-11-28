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
	minimum := dockWidths{dockProperties: 350, dockRightAuxiliary: 145}
	preferred := dockWidths{dockProperties: 454, dockRightAuxiliary: 330}
	got := fitDockWidths(space, 176, 265, minimum, preferred)
	if got[dockTools] != 0 || got[dockThumbnails] != 0 {
		t.Fatalf("the collapsed dock still takes space: %v", got)
	}
	if remaining := space - got[dockProperties] - got[dockRightAuxiliary]; remaining != 265 {
		t.Errorf("document width after panel growth = %d, want 265", remaining)
	}

	minimum = dockWidths{168, 350, 70, 145}
	preferred = dockWidths{195, 454, 220, 330}
	got = fitDockWidths(600, 176, 265, minimum, preferred)
	if got[dockThumbnails] != 0 || got[dockRightAuxiliary] != 0 || got[dockTools]+got[dockProperties] != 600-176 {
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

func TestSearchRestoresRightDock(t *testing.T) {
	for _, panel := range []auxiliaryPanel{auxiliaryNone, auxiliaryBookmarks, auxiliaryResults} {
		for _, collapsed := range []bool{false, true} {
			for _, primaryCollapsed := range []bool{false, true} {
				original := dockVisibility{collapsed: collapsed, primaryCollapsed: primaryCollapsed}
				dock, aux := original, rightAuxiliary{panel: panel}
				aux.syncSearch(true, &dock)
				if aux.panel != auxiliaryResults || !dock.showsPrimary() {
					t.Fatalf("search did not reveal its fields and results: %+v, panel %d", dock, aux.panel)
				}
				aux.syncSearch(true, &dock)
				aux.syncSearch(false, &dock)
				if dock != original || aux.panel != panel || aux.beforeSearch != nil {
					t.Fatalf("leaving search did not restore %+v, panel %d: got %+v, panel %d", original, panel, dock, aux.panel)
				}
			}
		}
	}
}

func TestSearchKeepsManualPanelChanges(t *testing.T) {
	tests := []struct {
		name  string
		edit  func(*rightAuxiliary, *dockVisibility)
		panel auxiliaryPanel
		dock  dockVisibility
	}{
		{"bookmarks then results", func(aux *rightAuxiliary, dock *dockVisibility) {
			aux.togglePanel(auxiliaryBookmarks)
			aux.togglePanel(auxiliaryResults)
		}, auxiliaryResults, dockVisibility{}},
		{"close results", func(aux *rightAuxiliary, dock *dockVisibility) {
			aux.togglePanel(auxiliaryResults)
		}, auxiliaryNone, dockVisibility{}},
		{"change the whole dock", func(aux *rightAuxiliary, dock *dockVisibility) {
			aux.setGroupCollapsed(dock, true)
		}, auxiliaryBookmarks, dockVisibility{collapsed: true}},
		{"change properties only", func(aux *rightAuxiliary, dock *dockVisibility) {
			aux.setPropertiesCollapsed(dock, true)
		}, auxiliaryBookmarks, dockVisibility{primaryCollapsed: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dock, aux := dockVisibility{}, rightAuxiliary{panel: auxiliaryBookmarks}
			aux.syncSearch(true, &dock)
			tt.edit(&aux, &dock)
			aux.syncSearch(true, &dock)
			aux.syncSearch(false, &dock)
			if dock != tt.dock || aux.panel != tt.panel {
				t.Fatalf("manual choice was lost: got %+v, panel %d, want %+v, panel %d", dock, aux.panel, tt.dock, tt.panel)
			}
		})
	}
}

func TestTemporaryDockVisibilityPreservesState(t *testing.T) {
	for bits := 0; bits < 8; bits++ {
		original := dockVisibility{
			collapsed: bits&1 != 0,
			primaryCollapsed: bits&2 != 0,
			auxiliaryCollapsed: bits&4 != 0,
		}
		hidden := temporaryDockVisibility(original, true)
		if hidden.showsPrimary() || hidden.showsAuxiliary(true) {
			t.Fatalf("temporary collapse leaked a panel: %+v", hidden)
		}
		if hidden.primaryCollapsed != original.primaryCollapsed || hidden.auxiliaryCollapsed != original.auxiliaryCollapsed {
			t.Fatalf("individual visibility changed: %+v -> %+v", original, hidden)
		}
		if got := temporaryDockVisibility(original, false); got != original {
			t.Fatalf("restore changed %+v to %+v", original, got)
		}
	}
}

func TestTemporaryDockVisibilityDoesNotMutateSearchSnapshot(t *testing.T) {
	original := dockVisibility{collapsed: true, auxiliaryCollapsed: true}
	dock := original
	aux := rightAuxiliary{panel: auxiliaryBookmarks}
	aux.syncSearch(true, &dock)
	searchDock := dock
	hidden := temporaryDockVisibility(dock, true)
	if !hidden.collapsed || dock != searchDock || aux.beforeSearch.dock != original {
		t.Fatal("temporary collapse changed search state")
	}
	aux.syncSearch(false, &dock)
	if dock != original || aux.panel != auxiliaryBookmarks {
		t.Fatal("search restoration failed after temporary collapse")
	}
}
