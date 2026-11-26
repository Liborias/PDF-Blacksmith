package main

import "math"

type cropMotionMode int

const (
	moveCropPair cropMotionMode = iota
	moveCropMask
	moveCropImage
)

type cropShifts struct {
	mask, image textPoint
	zoom        float64
}

func (s cropShifts) changed() bool { return s != (cropShifts{}) }

func (s cropShifts) imageScale() float64 { return 1 + s.zoom }

func (s cropShifts) imagePoint(p textPoint) textPoint {
	return sumCropPoints(s.image, scaleCropPoint(p, s.imageScale()))
}

func (s cropShifts) imageBounds(b searchRect) searchRect {
	return rectangleBetween(s.imagePoint(textPoint{b[0], b[1]}), s.imagePoint(textPoint{b[2], b[3]}))
}

func scaleCropPoint(p textPoint, factor float64) textPoint {
	return textPoint{p[0] * factor, p[1] * factor}
}

func (s cropShifts) scaledImage(quad, mask []textPoint, pivot textPoint, steps float64) cropShifts {
	if len(quad) != 4 || len(mask) < 3 {
		return s
	}
	scale := s.imageScale()
	change := min(8/scale, max(0.125/scale, math.Pow(1.125, steps))) - 1
	factor := 1 + limitCropEdges(quad, mask, func(p, normal textPoint) float64 {
		return change * cropDot(normal, cropDifference(pivot, p))
	})*change
	if math.Abs(factor-1) < 1e-7 {
		return s
	}
	s.image = sumCropPoints(pivot, scaleCropPoint(cropDifference(s.image, pivot), factor))
	s.zoom = min(8, max(.125, scale*factor)) - 1
	if math.Abs(s.zoom) < 1e-10 {
		s.zoom = 0
	}
	for axis := range s.image {
		if math.Abs(s.image[axis]) < 1e-10 {
			s.image[axis] = 0
		}
	}
	return s
}

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
	relative := cropDifference(shifts.mask, shifts.image)
	// Work in the image's original coordinate frame.
	current := make([]textPoint, len(quad))
	for i, p := range quad {
		current[i] = scaleCropPoint(p, shifts.imageScale())
	}
	outline := make([]textPoint, len(clip))
	for i, p := range clip {
		outline[i] = sumCropPoints(p, relative)
	}
	speed := desired
	if mode == moveCropImage {
		speed = scaleCropPoint(speed, -1)
	}
	amount = min(amount, limitCropEdges(current, outline, func(_, normal textPoint) float64 {
		return cropDot(normal, speed)
	}))
	return scaleCropPoint(desired, amount)
}

func cropDot(a, b textPoint) float64 { return a[0]*b[0] + a[1]*b[1] }

func limitCropEdges(quad, clip []textPoint, rate func(p, normal textPoint) float64) float64 {
	amount, orientation := 1., 1.
	a, b := cropDifference(quad[1], quad[0]), cropDifference(quad[3], quad[0])
	if a[0]*b[1]-a[1]*b[0] < 0 {
		orientation = -1
	}
	for i, p := range quad {
		q := quad[(i+1)%len(quad)]
		normal := textPoint{(p[1] - q[1]) * orientation, (q[0] - p[0]) * orientation}
		if speed := rate(p, normal); speed < 0 {
			distance := cropDot(normal, cropDifference(clip[0], p))
			for _, vertex := range clip[1:] {
				distance = min(distance, cropDot(normal, cropDifference(vertex, p)))
			}
			// Moving toward the interior can reduce an existing gap.
			amount = min(amount, max(0, distance)/-speed)
		}
	}
	return amount
}
