package main

import (
	"math"
	"slices"
)

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

func textParagraph(bounds []searchRect, seed int) []int {
	chosen := map[int]bool{seed: true}
	queue := []int{seed}
	margin := bounds[seed][0]
	for i := 0; i < len(queue) && len(queue) <= 64; i++ {
		a := bounds[queue[i]]
		height := a[3] - a[1]
		for j, b := range bounds {
			if chosen[j] || b[2] <= b[0] || b[3] <= b[1] {
				continue
			}
			other := b[3] - b[1]
			step := math.Abs((b[1] + b[3] - a[1] - a[3]) / 2)
			overlap := min(a[2], b[2]) - max(a[0], b[0])
			if other >= height*0.66 && other <= height*1.5 && step >= height*0.6 && step <= height*2.25 &&
				overlap > 0 && math.Abs(b[0]-margin) <= max(0.016, height*1.5) {
				chosen[j] = true
				queue = append(queue, j)
			}
		}
	}
	sortTextSelection(queue, bounds)
	return queue
}

func sortTextSelection(selected []int, bounds []searchRect) {
	slices.SortStableFunc(selected, func(i, j int) int {
		a, b := bounds[i], bounds[j]
		if a[1] < b[1] || (a[1] == b[1] && a[0] < b[0]) {
			return -1
		}
		if a[1] > b[1] || a[0] > b[0] {
			return 1
		}
		return 0
	})
}
