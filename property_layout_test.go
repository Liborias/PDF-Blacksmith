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


func TestCompactPropertyHeadersKeepControlsAndReports(t *testing.T) {
	items := []panelItem{
		{handle:1,x:16,y:16,height:22},
		{handle:2,x:290,y:16,height:22},
		{handle:3,x:16,y:110,height:30,section:1},
		{handle:4,x:16,y:210,height:22},
		{handle:5,x:290,y:210,height:22},
		{handle:6,x:16,y:350,height:59},
	}
	compactPropertyRowsWithHeaders(items,map[uintptr]bool{1:true,2:true,4:true,5:true})
	want:=[]int{16,16,48,88,88,120}
	for i,item:=range items {
		if item.y!=want[i] || item.handle!=uintptr(i+1) { t.Fatalf("item %d: %+v",i,item) }
	}
	if items[2].section!=1 || items[2].height!=30 || items[5].height!=59 { t.Fatal("control or report changed") }
}
