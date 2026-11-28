package main

import "testing"

func TestInitialPageImageReady(t *testing.T) {
	order := []int{1, 2, 3}
	for _, tc := range []struct {
		name string
		pending, secondary, ready bool
		current, rendered int
		want bool
	}{
		{"loading", true,false,false,0,0,false},
		{"description only",true,false,true,0,0,false},
		{"secondary render",true,true,true,0,1,false},
		{"other page",true,false,true,0,2,false},
		{"manual tool selection",false,false,true,0,1,false},
		{"invalid page index",true,false,true,3,1,false},
		{"first image",true,false,true,0,1,true},
		{"different current page",true,false,true,1,2,true},
	} {
		t.Run(tc.name,func(t *testing.T){
			if got:=initialPageImageReady(tc.pending,tc.secondary,tc.ready,order,tc.current,tc.rendered);got!=tc.want{
				t.Fatalf("got %v, want %v",got,tc.want)
			}
		})
	}
}
