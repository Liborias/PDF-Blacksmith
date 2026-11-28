package main

import "testing"

func TestPanelScrollThumbGeometry(t *testing.T) {
	for _, dpi := range []int{96,120,144,192} {
		d := func(n int) int { return n*dpi/96 }
		for _, h := range []int{1,2,10,20,28,40,70,120,250,600,1200} {
			height := d(h)
			for _, total := range []int{1,2,10,50,100,1000,10000} {
				for _, visible := range []int{1,2,5,20,100,1000} {
					last := max(0,total-visible)
					old := -1
					for _, position := range []int{-1,0,last/4,last/2,last,last+1} {
						offset,length := panelScrollThumb(height,d(28),total,visible,position)
						if offset<0 || length<1 || offset+length>height || offset<old {
							t.Fatalf("invalid geometry h=%d total=%d visible=%d pos=%d offset=%d length=%d",height,total,visible,position,offset,length)
						}
						if total<=visible && (offset!=0 || length!=height) { t.Fatal("non-scrollable track") }
						if total>visible && height>1 {
							if length>=height || length<min(d(28),height-1) { t.Fatal("invalid thumb minimum or travel") }
							if position>=last && offset+length!=height { t.Fatal("cannot reach bottom") }
							if position<=0 && offset!=0 { t.Fatal("cannot reach top") }
						}
						old=offset
					}
				}
			}
		}
	}
}
