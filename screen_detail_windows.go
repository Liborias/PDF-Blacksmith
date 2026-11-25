package main

import (
	"math"
	"slices"
)

const (
	screenRegionEdge   = 4096
	screenRegionPixels = 4 << 20
	screenRegionMargin = 64
	detailCacheCeiling = 64 << 20
)

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

type detailFrame struct {
	image *printBitmap
	used  uint64
}

func (v *documentView) pageBounds(page sheet) rect {
	x := int(v.viewport.left) + int(math.Round(page.left-v.scrollX))
	y := int(v.viewport.top) + int(math.Round(page.top-v.scrollY))
	width, height := max(1, int(math.Round(page.width))), max(1, int(math.Round(page.height)))
	return rect{int32(x), int32(y), int32(x + width), int32(y + height)}
}

func (v *documentView) clearDetails() {
	v.visibleDetails = nil
	v.detailFrames = make(map[screenTile]detailFrame)
	v.failedDetails = make(map[screenTile]bool)
	v.detailBytes = 0
}

// A zero target means the visible region needs a raster at the screen's scale.
func (v *documentView) fullPreviewTarget(page sheet) int {
	if v.proofMode() {
		return 1850
	}
	extent := max(page.width, page.height)
	if extent > 1420 {
		return 0
	}
	target := previewPixels(extent)
	frame := v.reading.frames[v.order[page.index]]
	if frame.pixels >= target && frame.image != nil &&
		(page.width > float64(frame.image.width+4) || page.height > float64(frame.image.height+4)) {
		return 0
	}
	return target
}

func (outer screenTile) covers(inner screenTile) bool {
	return outer.page == inner.page && outer.Zoom == inner.Zoom && outer.X <= inner.X && outer.Y <= inner.Y &&
		outer.X+outer.Width >= inner.X+inner.Width && outer.Y+outer.Height >= inner.Y+inner.Height
}

func (v *documentView) detailFor(tile screenTile, width, height int) screenTile {
	// During a pan, reuse a completed crop or a crop already requested for
	// this scale if its margin still covers the new visible area.
	for _, previous := range v.visibleDetails {
		if previous.covers(tile) {
			return previous
		}
	}
	var matching screenTile
	var latest uint64
	for existing, frame := range v.detailFrames {
		if existing.covers(tile) && (matching.page == 0 || frame.used > latest) {
			matching, latest = existing, frame.used
		}
	}
	if matching.page != 0 {
		return matching
	}
	padded := tile
	padded.X = max(0, tile.X-screenRegionMargin)
	padded.Y = max(0, tile.Y-screenRegionMargin)
	padded.Width = min(width, tile.X+tile.Width+screenRegionMargin) - padded.X
	padded.Height = min(height, tile.Y+tile.Height+screenRegionMargin) - padded.Y
	if padded.Width <= screenRegionEdge && padded.Height <= screenRegionEdge &&
		padded.Width*padded.Height <= screenRegionPixels {
		return padded
	}
	return tile
}

func (v *documentView) syncDetails() {
	var visible []screenTile
	if v.ready() && !v.overview && !v.proofMode() {
		_, viewportHeight := v.viewportSize()
		for _, page := range v.spread.visible(v.scrollY, viewportHeight) {
			if v.fullPreviewTarget(page) != 0 {
				continue
			}
			number := v.order[page.index]
			bounds := v.pageBounds(page)
			width, height := int(bounds.right-bounds.left), int(bounds.bottom-bounds.top)
			left, top := max(0, int(v.viewport.left-bounds.left)), max(0, int(v.viewport.top-bounds.top))
			right, bottom := min(width, int(v.viewport.right-bounds.left)), min(height, int(v.viewport.bottom-bounds.top))
			if left >= right || top >= bottom {
				continue
			}
			// Render a single viewport crop on ordinary displays. Very large
			// viewports are divided to bound the allocation of any request.
			chunkWidth := min(right-left, screenRegionEdge)
			chunkHeight := min(screenRegionEdge, screenRegionPixels/chunkWidth)
			for y := top; y < bottom; y += chunkHeight {
				for x := left; x < right; x += chunkWidth {
					needed := screenTile{number, screenRegion{v.zoom, x, y, min(right-x, chunkWidth), min(bottom-y, chunkHeight)}}
					tile := v.detailFor(needed, width, height)
					if !slices.Contains(visible, tile) {
						visible = append(visible, tile)
					}
				}
			}
		}
	}
	v.visibleDetails = visible
	v.detailClock++
	for _, tile := range visible {
		if frame, ok := v.detailFrames[tile]; ok {
			frame.used = v.detailClock
			v.detailFrames[tile] = frame
		}
	}
	for tile := range v.failedDetails {
		if !slices.Contains(visible, tile) {
			delete(v.failedDetails, tile)
		}
	}
}

func (v *documentView) trimDetails(budget int64) {
	for v.detailBytes > budget {
		var oldest screenTile
		var used uint64
		for tile, frame := range v.detailFrames {
			if !slices.Contains(v.visibleDetails, tile) && (oldest.page == 0 || frame.used < used) {
				oldest, used = tile, frame.used
			}
		}
		if oldest.page == 0 {
			return
		}
		v.detailBytes -= int64(len(v.detailFrames[oldest].image.pixels))
		delete(v.detailFrames, oldest)
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
			if v.secondary {
				v.showComparisonReaderError(job.result.Message)
			} else {
				v.app.updateActions()
				v.app.report(job.result.Message)
			}
		} else if job.frame.image != nil {
			v.detailFrames[tile] = detailFrame{job.frame.image, v.detailClock}
			v.detailBytes += int64(len(job.frame.image.pixels))
			v.trimCaches()
		} else {
			v.failedDetails[tile] = true
			if !v.secondary {
				v.app.report(job.result.Message)
			}
		}
		v.invalidateCanvas()
	}
	v.schedule()
}

func (v *documentView) paintDetails(dc uintptr, number int, bounds rect) {
	for _, tile := range v.visibleDetails {
		if bitmap := v.detailFrames[tile].image; tile.page == number && bitmap != nil {
			drawBitmap(dc, bitmap, int(bounds.left)+tile.X, int(bounds.top)+tile.Y, bitmap.width, bitmap.height)
		}
	}
}
