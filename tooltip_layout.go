package main

type tooltipBox struct { left,top,right,bottom int }

// side 0 prefers above; side +/-1 prefers the document-facing side of an ear.
// Never deliberately cover the anchor; leave the native fallback for an
// oversized popup or a screen with no fitting adjacent position.
func iconTooltipPosition(a,work tooltipBox,w,h,gap,side int) (int,int,bool) {
	if w<=0 || h<=0 || w>work.right-work.left || h>work.bottom-work.top { return 0,0,false }
	cx := (a.left+a.right-w)/2
	cy := (a.top+a.bottom-h)/2
	above,below := [2]int{cx,a.top-gap-h},[2]int{cx,a.bottom+gap}
	left,right := [2]int{a.left-gap-w,cy},[2]int{a.right+gap,cy}
	candidates := [][2]int{above,below,right,left}
	if side>0 { candidates=[][2]int{right,left,above,below} }
	if side<0 { candidates=[][2]int{left,right,above,below} }
	for _, p := range candidates {
		x:=max(work.left,min(p[0],work.right-w))
		y:=max(work.top,min(p[1],work.bottom-h))
		if x+w<=a.left || x>=a.right || y+h<=a.top || y>=a.bottom {
			return x,y,true
		}
	}
	return 0,0,false
}
