package main

import (
	"math"
	"slices"
	"testing"
)

func TestLayoutGridKeepsPhysicalSpacingAndOriginWhenClipped(t *testing.T) {
	type line struct {
		position int32
		major    bool
	}
	var got []line
	const pagePoints = 2 * layoutGridMajorPoints
	visitLayoutGridLines(10, 110, 36, 83, pagePoints, 5, func(position int32, major bool) {
		got = append(got, line{position, major})
	})
	want := []line{{40, false}, {50, false}, {60, true}, {70, false}, {80, false}}
	if !slices.Equal(got, want) {
		t.Errorf("clipped grid = %v, want %v", got, want)
	}
}

func TestLayoutGridReducesDensityWithoutDrawingPageBorder(t *testing.T) {
	var got []int32
	const pagePoints = 20 * layoutGridMajorPoints
	visitLayoutGridLines(-4, 36, -10, 40, pagePoints, 5, func(position int32, major bool) {
		if !major {
			t.Error("minor line at low zoom")
		}
		got = append(got, position)
	})
	if want := []int32{2, 8, 14, 20, 26, 32}; !slices.Equal(got, want) {
		t.Errorf("low-zoom grid = %v, want %v", got, want)
	}
	for _, invalid := range []float64{0, math.NaN(), math.Inf(1)} {
		visitLayoutGridLines(0, 100, 0, 100, invalid, 5, func(position int32, _ bool) {
			t.Errorf("page size %f produced line at %d", invalid, position)
		})
	}
}
