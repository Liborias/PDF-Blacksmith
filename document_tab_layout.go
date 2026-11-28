package main

type documentTabLabelPlan struct {
	start,end,dot int
	icon bool
}

// Reserve the dirty indicator independently of the caption. At small
// widths the optional file icon yields before the filename.
func documentTabLabelLayout(width,pad,icon,gap,dot,minText int,dirty bool) documentTabLabelPlan {
	width=max(0,width)
	pad=min(max(0,pad),width/2)
	p:=documentTabLabelPlan{start:pad,end:width-pad,dot:-1}
	if dirty && p.end-p.start>=dot+gap {
		p.dot=p.end-dot
		p.end=p.dot-gap
	}
	if p.end-p.start>=icon+gap+minText {
		p.icon=true
		p.start+=icon+gap
	}
	return p
}
