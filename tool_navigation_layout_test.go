package main

import "testing"

func TestToolNavigationFitsAvailableHeight(t *testing.T) {
	for _,dpi:=range []int{96,120,144,192} {
		d:=func(n int) int { return n*dpi/96 }
		for _,groups:=range [][]int{{1,1,3},{1,1,2,2,2,2,2,2,3,3,3}} {
			for h:=200;h<=1600;h++ {
				height:=d(h)
				p:=planToolNavigation(height,groups,d(10),d(5),d(39),d(27),d(25),d(7))
				last:=0
				for _,row:=range p.rows {
					if row.top<last || row.height<0 || row.top+row.height>height {
						t.Fatalf("overflow h=%d dpi=%d plan=%+v",h,dpi,p)
					}
					if height>=2*d(10)+len(groups)*d(27) && row.height<d(27)-d(5) {
						t.Fatal("decorations displaced compact tool rows")
					}
					last=row.top+row.height
				}
				if len(p.separators)>len(groups)-1 { t.Fatal("orphan separator") }
				if h>=800 && (p.heading!=d(25)||len(p.separators)==0) { t.Fatal("wide layout did not restore decoration") }
			}
		}
	}
}
