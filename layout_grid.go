package main

import "math"

const layoutGridMajorPoints = 5 * 72.0 / 25.4

// Visit only the visible interior lines. At low zoom skip dense divisions
// while keeping the grid anchored to the page, even when the page is panned.
func visitLayoutGridLines(start, end, clipStart, clipEnd int32, pagePoints, minimumGap float64, visit func(int32, bool)) {
	if start >= end || pagePoints <= 0 {
		return
	}
	clipStart, clipEnd = max(start+1, clipStart), min(end, clipEnd)
	if clipStart >= clipEnd {
		return
	}
	origin := float64(start)
	step := (float64(end) - origin) * layoutGridMajorPoints / pagePoints
	if step <= 0 || math.IsNaN(step) || math.IsInf(step, 0) {
		return
	}
	minimumGap = max(1, minimumGap)
	majorEvery := 1
	if step/5 >= minimumGap {
		step /= 5
		majorEvery = 5
	} else if step < minimumGap {
		step *= math.Ceil(minimumGap / step)
	}
	first := max(1, int(math.Ceil((float64(clipStart)-origin-0.5)/step)))
	for index := first; ; index++ {
		position := origin + math.Round(float64(index)*step)
		if position >= float64(clipEnd) {
			break
		}
		visit(int32(position), index%majorEvery == 0)
	}
}
