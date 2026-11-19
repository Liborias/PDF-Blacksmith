package main

// A searchRect from PDFium is [left, bottom, right, top] in PDF units.
// The normalized result is [left, top, right, bottom] in display order.
type searchRect [4]float64

func searchDisplayRect(bounds, box searchRect, clockwise int) (searchRect, bool) {
	w, h := box[2]-box[0], box[3]-box[1]
	left, right := max(bounds[0], box[0]), min(bounds[2], box[2])
	bottom, top := max(bounds[1], box[1]), min(bounds[3], box[3])
	if w <= 0 || h <= 0 || right <= left || top <= bottom {
		return searchRect{}, false
	}
	x, y := (left-box[0])/w, (box[3]-top)/h
	r, b := (right-box[0])/w, (box[3]-bottom)/h
	switch clockwise % 360 {
	case 90:
		return searchRect{1 - b, x, 1 - y, r}, true
	case 180:
		return searchRect{1 - r, 1 - b, 1 - x, 1 - y}, true
	case 270:
		return searchRect{y, 1 - r, b, 1 - x}, true
	default:
		return searchRect{x, y, r, b}, true
	}
}
