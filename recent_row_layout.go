package main

// Return the two line rectangles relative to the row's top.
func recentRowLines(height,padding,gap int) (nameTop,nameBottom,pathTop,pathBottom int) {
	height=max(0,height)
	padding=min(max(0,padding),height/2)
	space:=height-2*padding
	gap=min(max(0,gap),space)
	nameTop=padding
	nameBottom=nameTop+(space-gap+1)/2
	pathTop=nameBottom+gap
	pathBottom=height-padding
	return
}
