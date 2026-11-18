package main

import (
	"math"
	"slices"
)

const screenTileSize = 1536

type screenRegion struct {
	Zoom   float64 `json:"zoom"`
	X      int     `json:"x"`
	Y      int     `json:"y"`
	Width  int     `json:"width"`
	Height int     `json:"height"`
}

type screenTile struct {
	page int
	screenRegion
}

func (v *documentView) pageBounds(page sheet) rect {
	x := int(v.viewport.left) + int(math.Round(page.left-v.scrollX))
	y := int(v.viewport.top) + int(math.Round(page.top-v.scrollY))
	width, height := max(1, int(math.Round(page.width))), max(1, int(math.Round(page.height)))
	return rect{int32(x), int32(y), int32(x + width), int32(y + height)}
}

func (v *documentView) clearDetails() {
	v.visibleDetails = nil
	v.detailFrames = make(map[screenTile]*printBitmap)
	v.failedDetails = make(map[screenTile]bool)
	v.detailBytes = 0
}

func (v *documentView) syncDetails() {
	v.visibleDetails = v.visibleDetails[:0]
	if v.ready() && !v.proofMode() {
		_, height := v.viewportSize()
		for _, page := range v.spread.visible(v.scrollY, height) {
			number := v.order[page.index]
			frame := v.reading.frames[number]
			bounds := v.pageBounds(page)
			width, height := int(bounds.right-bounds.left), int(bounds.bottom-bounds.top)
			if frame.image == nil || (width <= frame.image.width+4 && height <= frame.image.height+4) {
				continue
			}
			left := max(0, int(v.viewport.left-bounds.left))
			top := max(0, int(v.viewport.top-bounds.top))
			right := min(width, int(v.viewport.right-bounds.left))
			bottom := min(height, int(v.viewport.bottom-bounds.top))
			if left >= right || top >= bottom {
				continue
			}
			for y := top / screenTileSize * screenTileSize; y < bottom; y += screenTileSize {
				for x := left / screenTileSize * screenTileSize; x < right; x += screenTileSize {
					v.visibleDetails = append(v.visibleDetails, screenTile{number, screenRegion{v.zoom, x, y,
						min(width-x, screenTileSize), min(height-y, screenTileSize)}})
				}
			}
		}
	}
	for tile, frame := range v.detailFrames {
		if !slices.Contains(v.visibleDetails, tile) {
			v.detailBytes -= int64(len(frame.pixels))
			delete(v.detailFrames, tile)
		}
	}
	for tile := range v.failedDetails {
		if !slices.Contains(v.visibleDetails, tile) {
			delete(v.failedDetails, tile)
		}
	}
}

func (v *documentView) renderDetail(tile screenTile) {
	req := v.request("screen")
	req.View.Page, req.View.Rotation = tile.page, v.rotations[tile.page-1]
	req.View.Region = &tile.screenRegion
	v.launch(req, tile.page, 0, false, &tile)
}

func (v *documentView) finishDetail(job readerJob) {
	tile := *job.tile
	if slices.Contains(v.visibleDetails, tile) {
		if job.result.Status == "changed" {
			v.invalidSource = true
			v.app.updateActions()
			v.app.report(job.result.Message)
		} else if job.frame.image != nil {
			v.detailFrames[tile] = job.frame.image
			v.detailBytes += int64(len(job.frame.image.pixels))
			v.trimCaches()
		} else {
			v.failedDetails[tile] = true
			v.app.report(job.result.Message)
		}
		invalidate(v.app.window)
	}
	v.schedule()
}

func (v *documentView) paintDetails(dc uintptr, number int, bounds rect) {
	for _, tile := range v.visibleDetails {
		if bitmap := v.detailFrames[tile]; tile.page == number && bitmap != nil {
			drawBitmap(dc, bitmap, int(bounds.left)+tile.X, int(bounds.top)+tile.Y, bitmap.width, bitmap.height)
		}
	}
}
