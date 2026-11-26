package main

type cropMotionMode int

const (
	moveCropPair cropMotionMode = iota
	moveCropMask
	moveCropImage
)

type cropShifts struct {
	mask, image textPoint
}

func (s cropShifts) changed() bool { return s != (cropShifts{}) }

func sumCropPoints(a, b textPoint) textPoint  { return textPoint{a[0] + b[0], a[1] + b[1]} }
func cropDifference(a, b textPoint) textPoint { return textPoint{a[0] - b[0], a[1] - b[1]} }

func moveCropBounds(box searchRect, delta textPoint) searchRect {
	return searchRect{box[0] + delta[0], box[1] + delta[1], box[2] + delta[0], box[3] + delta[1]}
}

// Limit relative motion against each edge of the transformed raster. A legacy
// mask may already extend past an edge; allow it to move inward from there.
func constrainCropMotion(quad, clip []textPoint, shifts cropShifts, mode cropMotionMode, bounds searchRect, desired textPoint) textPoint {
	amount := 1.0
	if mode != moveCropImage {
		for axis := range 2 {
			switch {
			case desired[axis] < 0:
				amount = min(amount, max(0, bounds[axis])/-desired[axis])
			case desired[axis] > 0:
				amount = min(amount, max(0, 1-bounds[axis+2])/desired[axis])
			}
		}
	}
	if mode == moveCropPair {
		return textPoint{amount * desired[0], amount * desired[1]}
	}
	if len(quad) != 4 || len(clip) < 3 {
		return textPoint{}
	}
	orientation := 1.
	a, b := cropDifference(quad[1], quad[0]), cropDifference(quad[3], quad[0])
	if a[0]*b[1]-a[1]*b[0] < 0 {
		orientation = -1
	}
	relative := cropDifference(shifts.mask, shifts.image)
	for i, p := range quad {
		q := quad[(i+1)%len(quad)]
		normal := textPoint{(p[1] - q[1]) * orientation, (q[0] - p[0]) * orientation}
		dot := func(v textPoint) float64 { return normal[0]*v[0] + normal[1]*v[1] }
		edge := dot(cropDifference(clip[0], p))
		for _, vertex := range clip[1:] {
			edge = min(edge, dot(cropDifference(vertex, p)))
		}
		rate := dot(desired)
		if mode == moveCropImage {
			rate = -rate
		}
		if rate < 0 {
			room := max(0, edge+dot(relative)-min(0, edge))
			amount = min(amount, room/-rate)
		}
	}
	return textPoint{amount * desired[0], amount * desired[1]}
}
