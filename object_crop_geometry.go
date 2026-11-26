package main

import "math"

// Fit the dragged corner in display-oriented PDF units. Normalized coordinates
// have different scales on the two axes, including after a quarter turn.
func constrainImageCrop(start, end textPoint, image, page searchRect, rotation int, ratio float64) textPoint {
	if ratio <= 0 {
		return textPoint{min(image[2], max(image[0], end[0])), min(image[3], max(image[1], end[1]))}
	}
	width, height := page[2]-page[0], page[3]-page[1]
	if rotation%180 != 0 {
		width, height = height, width
	}
	aspect := ratio * height / width
	dx, dy := end[0]-start[0], end[1]-start[1]
	xDirection, xLimit := cropAxis(start[0], dx, image[0], image[2])
	yDirection, yLimit := cropAxis(start[1], dy, image[1], image[3])
	cropHeight := min(max(math.Abs(dx)/aspect, math.Abs(dy)), xLimit/aspect, yLimit)
	return textPoint{start[0] + xDirection*cropHeight*aspect, start[1] + yDirection*cropHeight}
}

func cropAxis(anchor, delta, low, high float64) (direction, room float64) {
	if delta < 0 || delta == 0 && anchor-low > high-anchor {
		return -1, anchor - low
	}
	return 1, high - anchor
}
