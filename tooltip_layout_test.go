package main

import "testing"

func TestIconTooltipPosition(t *testing.T) {
	work:=tooltipBox{-1920,0,0,1080}
	for _,side:=range []int{-1,0,1} {
		for _,x:=range []int{-1920,-1900,-1000,-100,-30} {
			for _,y:=range []int{0,15,50,400,1020,1050} {
				a:=tooltipBox{x,y,x+30,y+30}
				for _,w:=range []int{80,200,350} {
					for _,h:=range []int{24,50,110} {
						tx,ty,ok:=iconTooltipPosition(a,work,w,h,6,side)
						if !ok || tx<work.left || ty<work.top || tx+w>work.right || ty+h>work.bottom ||
							(tx+w>a.left && tx<a.right && ty+h>a.top && ty<a.bottom) {
							t.Fatalf("bad placement anchor=%+v side=%d x=%d y=%d",a,side,tx,ty)
						}
					}
				}
			}
		}
	}
	x,y,ok:=iconTooltipPosition(tooltipBox{-1000,50,-970,80},work,100,24,6,0)
	if !ok || x!=-1035 || y!=20 { t.Fatalf("above preference: %d %d %v",x,y,ok) }
}
