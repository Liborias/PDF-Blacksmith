package main

import (
	"math"
	"testing"
)

func TestImageCropRatioOnRotatedPages(t *testing.T) {
	page := searchRect{24, -30, 648, 762}
	image := searchRect{.15, .2, .85, .7}
	start := textPoint{.4, .45}
	ratios := []float64{1, 4. / 3, 3. / 4, 16. / 9, 9. / 16, 3. / 2, 2. / 3, 5. / 4, 4. / 5, 21. / 9}
	for _, angle := range []int{0, 90, 180, 270} {
		for _, ratio := range ratios {
			for _, pointer := range []textPoint{{.53, .54}, {.3, .35}, {.95, .95}, {.03, .03}, {.95, .03}, {.03, .95}} {
				end := constrainImageCrop(start, pointer, image, page, angle, ratio)
				if end[0] < image[0]-1e-12 || end[0] > image[2]+1e-12 ||
					end[1] < image[1]-1e-12 || end[1] > image[3]+1e-12 {
					t.Fatalf("angle %d, ratio %g: crop corner %v escaped image %v", angle, ratio, end, image)
				}
				if (pointer[0] < start[0]) != (end[0] < start[0]) ||
					(pointer[1] < start[1]) != (end[1] < start[1]) {
					t.Fatalf("angle %d, ratio %g: pointer %v and crop corner %v disagree about direction", angle, ratio, pointer, end)
				}
				pdfStart := textPointToPDF(start, page, angle)
				pdfEnd := textPointToPDF(end, page, angle)
				width, height := math.Abs(pdfEnd[0]-pdfStart[0]), math.Abs(pdfEnd[1]-pdfStart[1])
				if angle%180 != 0 {
					width, height = height, width
				}
				if math.Abs(width/height-ratio) > 1e-10 || height == 0 || width == 0 {
					t.Fatalf("angle %d, ratio %g: got crop dimensions %g × %g", angle, ratio, width, height)
				}
			}
		}
	}
}

func TestImageCropFreeAndSingleAxisDrag(t *testing.T) {
	page := searchRect{0, 0, 600, 900}
	image := searchRect{.1, .2, .9, .8}
	start := textPoint{.9, .4}
	if end := constrainImageCrop(start, textPoint{.95, .86}, image, page, 0, 0); end != (textPoint{.9, .8}) {
		t.Fatalf("free crop ended at %v, want image corner", end)
	}
	if end := constrainImageCrop(start, start, image, page, 0, 1); end != start {
		t.Fatalf("click grew a crop to %v", end)
	}
	if end := constrainImageCrop(start, textPoint{.9, .7}, image, page, 0, 1); end[0] >= start[0] || end[1] <= start[1] {
		t.Fatalf("vertical square drag at right edge ended at %v, want a crop down and to the left", end)
	}
}

func TestResizeImageCropHandles(t *testing.T) {
	limit := searchRect{0, 0, 1, 1}
	box := searchRect{.25, .25, .75, .75}
	minimum := textPoint{.01, .01}
	for _, test := range []struct {
		handle cropHandle
		delta  textPoint
		aspect float64
		want   searchRect
	}{
		{cropHandle{-1, -1}, textPoint{-.1, -.2}, 0, searchRect{.15, .05, .75, .75}},
		{cropHandle{0, -1}, textPoint{0, -.15}, 1, searchRect{.175, .1, .825, .75}},
		{cropHandle{1, 0}, textPoint{1, 0}, 1, searchRect{.25, .125, 1, .875}},
		{cropHandle{1, 1}, textPoint{-1, -1}, 0, searchRect{.25, .25, .26, .26}},
	} {
		got := resizeImageCrop(box, limit, test.handle, test.delta, minimum, test.aspect)
		for index := range got {
			if math.Abs(got[index]-test.want[index]) > 1e-7 {
				t.Errorf("resize %v: got %v, want %v", test.handle, got, test.want)
				break
			}
		}
	}
}

func TestCropMotionRespectsActualRasterEdges(t *testing.T) {
	quad := []textPoint{{.2, .1}, {.7, .1}, {.9, .9}, {.4, .9}}
	clip := []textPoint{{.45, .35}, {.65, .35}, {.65, .55}, {.45, .55}}
	bounds := polygonBounds(clip)
	for _, tc := range []struct {
		mode    cropMotionMode
		request textPoint
		want    textPoint
	}{
		{moveCropPair, textPoint{.8, 0}, textPoint{.35, 0}},
		{moveCropMask, textPoint{-.6, 0}, textPoint{-.1375, 0}},
		{moveCropImage, textPoint{.6, 0}, textPoint{.1375, 0}},
	} {
		got := constrainCropMotion(quad, clip, cropShifts{}, tc.mode, bounds, tc.request)
		for i := range got {
			if math.Abs(got[i]-tc.want[i]) > 1e-9 {
				t.Errorf("mode %v: got %v, want %v", tc.mode, got, tc.want)
				break
			}
		}
	}
}

func TestCropImageScaleAnchorsAndStopsAtImageEdge(t *testing.T) {
	quad := []textPoint{{.10, .1}, {.9, .1}, {.9, .9}, {.1, .9}}
	outline := []textPoint{{.25, .25}, {.6, .25}, {.6, .6}, {.25, .6}}
	pivot := textPoint{.4, .4}
	zoom := (cropShifts{}).scaledImage(quad, outline, pivot, 1)
	if got := zoom.imagePoint(pivot); got != pivot || zoom.imageScale() != 1.125 {
		t.Fatalf("pivot %v shifted to %v at %.3f", pivot, got, zoom.imageScale())
	}
	zoom = (cropShifts{}).scaledImage(quad, outline, pivot, -20)
	if s := zoom.imageScale(); math.Abs(s-.5) > 1e-10 {
		t.Fatalf("zoomed past the nearest image edge, scale %g", s)
	}
	for i := range quad {
		quad[i] = zoom.imagePoint(quad[i])
	}
	if again := zoom.scaledImage(quad, outline, pivot, -4); again != zoom {
		t.Fatalf("could still zoom out past the edge: %v", again)
	}
}
