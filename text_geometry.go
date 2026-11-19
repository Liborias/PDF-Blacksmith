package main

import "math"

type textPoint [2]float64

func rectangleSelection(a, b textPoint) []textPoint {
	return []textPoint{a, {b[0], a[1]}, b, {a[0], b[1]}}
}

func pointInTextRegion(point textPoint, polygon []textPoint) bool {
	if len(polygon) < 3 {
		return false
	}
	inside := false
	j := len(polygon) - 1
	for i, end := range polygon {
		start := polygon[j]
		if (end[1] > point[1]) != (start[1] > point[1]) &&
			point[0] < (start[0]-end[0])*(point[1]-end[1])/(start[1]-end[1])+end[0] {
			inside = !inside
		}
		j = i
	}
	return inside
}

func textPointToPDF(p textPoint, box searchRect, rotation int) textPoint {
	x, y := p[0], 1-p[1]
	switch rotation % 360 {
	case 90:
		x, y = p[1], p[0]
	case 180:
		x, y = 1-p[0], p[1]
	case 270:
		x, y = 1-p[1], 1-p[0]
	}
	return textPoint{box[0] + x*(box[2]-box[0]), box[1] + y*(box[3]-box[1])}
}

func textHit(bounds []searchRect, p textPoint, tolerance textPoint) int {
	found, distance := -1, math.Inf(1)
	for i, b := range bounds {
		w, h := b[2]-b[0], b[3]-b[1]
		if w <= 0 || h <= 0 || p[0] < b[0]-tolerance[0] || p[0] > b[2]+tolerance[0] ||
			p[1] < b[1]-tolerance[1] || p[1] > b[3]+tolerance[1] {
			continue
		}
		edge := max(0, b[0]-p[0], p[0]-b[2]) + max(0, b[1]-p[1], p[1]-b[3])
		score := edge*2 + w*h
		if score < distance {
			found, distance = i, score
		}
	}
	return found
}

func pdfPointToText(p textPoint, box searchRect, rotation int) textPoint {
	x, y := (p[0]-box[0])/(box[2]-box[0]), (box[3]-p[1])/(box[3]-box[1])
	switch rotation % 360 {
	case 90:
		return textPoint{1 - y, x}
	case 180:
		return textPoint{1 - x, 1 - y}
	case 270:
		return textPoint{y, 1 - x}
	}
	return textPoint{x, y}
}

type selectionChange uint8

const (
	replaceSelection selectionChange = iota
	addSelection
	toggleSelection
	subtractSelection
)

func combineSelection[T any, K comparable](current, picked []T, change selectionChange, key func(T) K) []T {
	if change == replaceSelection {
		return picked
	}
	previous := make(map[K]bool, len(current))
	hits := make(map[K]bool, len(picked))
	for _, entry := range picked {
		hits[key(entry)] = true
	}
	result := make([]T, 0, len(current)+len(picked))
	for _, entry := range current {
		id := key(entry)
		previous[id] = true
		if change == addSelection || !hits[id] {
			result = append(result, entry)
		}
	}
	if change != subtractSelection {
		for _, entry := range picked {
			if !previous[key(entry)] {
				result = append(result, entry)
			}
		}
	}
	return result
}

func rectangleBetween(first, last textPoint) searchRect {
	return searchRect{min(first[0], last[0]), min(first[1], last[1]), max(first[0], last[0]), max(first[1], last[1])}
}

func boundsContain(outer, inner searchRect) bool {
	return outer[0] <= inner[0] && outer[1] <= inner[1] && outer[2] >= inner[2] && outer[3] >= inner[3]
}

func rectangleSelects(selection, bounds searchRect, crossing bool) bool {
	if !crossing {
		return boundsContain(selection, bounds)
	}
	return selection[0] <= bounds[2] && selection[2] >= bounds[0] &&
		selection[1] <= bounds[3] && selection[3] >= bounds[1]
}

func rectangleTouchesSegment(box searchRect, a, b textPoint) bool {
	first, last := 0., 1.
	for axis := range 2 {
		delta := b[axis] - a[axis]
		if delta == 0 {
			if a[axis] < box[axis] || a[axis] > box[axis+2] {
				return false
			}
			continue
		}
		lo := (box[axis] - a[axis]) / delta
		hi := (box[axis+2] - a[axis]) / delta
		first, last = max(first, min(lo, hi)), min(last, max(lo, hi))
		if first > last {
			return false
		}
	}
	return true
}
