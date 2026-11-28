package main

import "testing"

func TestFileMenuHoverGrace(t *testing.T) {
	ticks:=0
	for i,inside:=range []bool{true,false,true,false,false,true} {
		ticks=fileMenuOutsideTicks(ticks,inside)
		if ticks!=[]int{0,1,0,1,2,0}[i] { t.Fatalf("step %d: %d",i,ticks) }
	}
}
