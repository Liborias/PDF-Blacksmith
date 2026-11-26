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
