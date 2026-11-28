package main

// Match the reference minimum thumb length while retaining at least one
// pixel of travel for short, scrollable panels. Offset and drag use the
// same travel distance, so both endpoints remain reachable.
func panelScrollThumb(height, minThumb, total, visible, position int) (offset, length int) {
	height = max(1,height)
	if total <= visible || total <= 0 {
		return 0,height
	}
	length = min(height,max(minThumb,height*max(0,visible)/total))
	if height > 1 {
		length = min(length,height-1)
	}
	last := total-visible
	position = min(last,max(0,position))
	return (height-length)*position/last,length
}
