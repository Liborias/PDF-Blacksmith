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

// A handle has a moving edge on each nonzero axis. A locked ratio
// keeps the opposite edge fixed and grows symmetrically on the other axis.
type cropHandle [2]int

var cropHandles = [...]cropHandle{{-1, -1}, {0, -1}, {1, -1}, {1, 0}, {1, 1}, {0, 1}, {-1, 1}, {-1, 0}}

func (handle cropHandle) point(box searchRect) textPoint {
	return textPoint{
		box[0] + float64(handle[0]+1)*(box[2]-box[0])/2,
		box[1] + float64(handle[1]+1)*(box[3]-box[1])/2,
	}
}

func resizeImageCrop(box, limit searchRect, handle cropHandle, delta, minimum textPoint, aspect float64) searchRect {
	anchor := cropHandle{-handle[0], -handle[1]}.point(box)
	cursor := handle.point(box)
	extent, room := textPoint{}, textPoint{}
	for axis := range 2 {
		direction := float64(handle[axis])
		if direction == 0 {
			extent[axis] = box[axis+2] - box[axis]
			room[axis] = 2 * min(anchor[axis]-limit[axis], limit[axis+2]-anchor[axis])
			continue
		}
		room[axis] = limit[axis+2] - anchor[axis]
		if direction < 0 {
			room[axis] = anchor[axis] - limit[axis]
		}
		extent[axis] = min(room[axis], max(minimum[axis], direction*(cursor[axis]+delta[axis]-anchor[axis])))
	}
	if aspect > 0 {
		height := max(extent[1], extent[0]/aspect)
		switch {
		case handle[0] == 0:
			height = extent[1]
		case handle[1] == 0:
			height = extent[0] / aspect
		}
		height = min(room[1], room[0]/aspect, max(height, minimum[1], minimum[0]/aspect))
		extent = textPoint{height * aspect, height}
	}
	for axis, direction := range handle {
		switch direction {
		case -1:
			box[axis], box[axis+2] = anchor[axis]-extent[axis], anchor[axis]
		case 1:
			box[axis], box[axis+2] = anchor[axis], anchor[axis]+extent[axis]
		default:
			box[axis], box[axis+2] = anchor[axis]-extent[axis]/2, anchor[axis]+extent[axis]/2
		}
	}
	return box
}
