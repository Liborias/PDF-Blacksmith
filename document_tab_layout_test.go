package main

import "testing"

func TestDocumentTabLabelLayout(t *testing.T) {
	for _,dpi:=range []int{96,120,144,192} {
		d:=func(n int) int { return n*dpi/96 }
		for width:=0;width<=d(240);width++ {
			for _,dirty:=range []bool{false,true} {
				p:=documentTabLabelLayout(width,d(7),d(16),d(5),d(7),d(24),dirty)
				if p.start<0 || p.end<p.start || p.end>width { t.Fatalf("label overflow: %+v",p) }
				if p.dot>=0 && (!dirty || p.dot<p.end || p.dot+d(7)>width) { t.Fatalf("dirty overlap: %+v",p) }
				if width>=d(59) && dirty && p.dot<0 { t.Fatal("missing dirty mark") }
				if p.icon && p.end-p.start<d(24) { t.Fatal("icon displaced caption") }
			}
		}
	}
}
