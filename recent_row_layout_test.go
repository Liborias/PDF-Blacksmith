package main

import "testing"

func TestRecentRowLines(t *testing.T) {
	for _,dpi:=range []int{96,120,144,192} {
		d:=func(n int) int { return n*dpi/96 }
		for height:=0;height<=d(60);height++ {
			nt,nb,pt,pb:=recentRowLines(height,d(3),d(2))
			if nt<0 || nb<nt || pt<nb || pb<pt || pb>height {
				t.Fatalf("bad row geometry DPI=%d height=%d: %d %d %d %d",dpi,height,nt,nb,pt,pb)
			}
			if height==d(44) && (nb-nt<d(16)||pb-pt<d(16)) { t.Fatal("two lines do not fit") }
		}
	}
}
