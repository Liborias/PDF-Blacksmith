package main

import "testing"

func TestFooterSlotsNeverOverlap(t *testing.T) {
	for _, dpi := range []int{96,120,144,192} {
		d := func(n int) int { return n*dpi/96 }
		for _, mode := range []int{96,120,145} {
			for space := 0; space <= 1800; space++ {
				a := d(space)
				p := fitFooterSlots(a,d(8),d(mode),d(120),d(172),d(70),d(140),d(218))
				used,count := 0,0
				for _, w := range []int{p.mode,p.dirty,p.tool,p.help,p.status} {
					if w>0 { used+=w; count++ }
				}
				used+=max(0,count-1)*d(8)
				if used>a || p.status<min(a,d(140)) {
					t.Fatalf("overflow or lost status: space=%d dpi=%d plan=%+v",space,dpi,p)
				}
				if a>=d(140)+d(70)+d(8) && p.help<d(70) {
					t.Fatal("lost existing help slot")
				}
				if space>=900 && (p.mode==0 || p.dirty==0 || p.tool==0) {
					t.Fatal("wide layout did not restore metadata")
				}
			}
		}
	}
}
