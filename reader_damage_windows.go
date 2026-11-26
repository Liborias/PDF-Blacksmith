package main

import (
	"math"
	"slices"
)

// Damage is expressed in the selected page box, after display rotation.
type pageDamage struct {
	page int
	box  string
	area searchRect
}

func (v *documentView) isVersion(base documentContent) bool {
	return v.path == base.path && v.draft == base.draft &&
		slices.Equal(v.order, base.order) && slices.Equal(v.rotations, base.rotations)
}

func (a *app) swapLocalVersion(next documentContent, pages []int, damage *pageDamage, visuallyUnchanged bool) {
	v := a.view
	query := getText(v.search.input)
	v.resetSearch()
	a.clearTextSelection()
	move := &v.contentMove
	if move.dropped && move.backdrop != nil && damage != nil && move.pick.source == damage.page {
		move.settling = true
	} else {
		v.cancelCanvasContentMove()
	}
	v.clearCanvasContent()
	v.cancelCanvasSelection()
	a.clearObjectSelection()
	v.selectionHistory = canvasSelectionHistory{}
	v.nativeText, v.objectProbe = textProbeState{}, objectProbeState{}
	v.generation++
	// Geometry, page order and bookmarks match. Keep the camera, selection of
	// pages and outstanding caches while switching the file revision.
	v.draft, v.draftInfo, v.history = next.draft, next.draftInfo, next.history
	if !visuallyUnchanged {
		for _, page := range pages {
			v.invalidateEditedPage(page, damage)
		}
		invalidate(v.pageList)
	}
	setText(v.search.input, query)
	v.updateDocumentInfo()
	a.updateActions()
	v.changedViewport()
	v.settleContentPreview(v.contentMove.pick.source)
}

func (v *documentView) invalidateEditedPage(page int, damage *pageDamage) {
	// Thumbnails and print proof may use a different page box or color model.
	v.thumbnails.remove(page)
	v.printPreview.remove(page)
	delete(v.prefetchFailed, page)
	if v.proofPage == page {
		v.clearPlateView()
	}

	repair := damage != nil && damage.page == page && damage.box == v.box() && !v.proofMode()
	base := screenTile{page: page}
	if frame := v.reading.frames[page]; repair && frame.image != nil && frame.scale > 0 {
		v.markRasterPatch(base, damage.area)
	} else {
		delete(v.rasterDamage, base)
		v.reading.remove(page)
	}
	for tile, frame := range v.detailFrames {
		if tile.page != page {
			continue
		}
		if repair {
			if _, needed := v.patchRegion(tile, damage.area); needed {
				v.markRasterPatch(tile, damage.area)
			}
		} else {
			v.detailBytes -= int64(len(frame.image.pixels))
			delete(v.detailFrames, tile)
			delete(v.rasterDamage, tile)
		}
	}
	for tile := range v.failedDetails {
		if tile.page == page {
			delete(v.failedDetails, tile)
		}
	}
}

func unionPageRects(a, b searchRect) searchRect {
	if a == (searchRect{}) {
		return b
	}
	return searchRect{min(a[0], b[0]), min(a[1], b[1]), max(a[2], b[2]), max(a[3], b[3])}
}

func (v *documentView) markRasterPatch(tile screenTile, area searchRect) {
	v.rasterDamage[tile] = unionPageRects(v.rasterDamage[tile], area)
}

func (v *documentView) patchRegion(tile screenTile, area searchRect) (screenRegion, bool) {
	region := tile.screenRegion
	size := v.pageDimensions(tile.page, v.box())
	if region.Zoom == 0 {
		frame := v.reading.frames[tile.page]
		if frame.image == nil || frame.scale <= 0 {
			return screenRegion{}, false
		}
		region = screenRegion{Zoom: frame.scale, Width: frame.image.width, Height: frame.image.height}
	}
	w, h := math.Round(size.width*region.Zoom), math.Round(size.height*region.Zoom)
	// A few pixels on each side also cover edge antialiasing.
	x := max(region.X, int(math.Floor(area[0]*w))-3)
	y := max(region.Y, int(math.Floor(area[1]*h))-3)
	right := min(region.X+region.Width, int(math.Ceil(area[2]*w))+3)
	bottom := min(region.Y+region.Height, int(math.Ceil(area[3]*h))+3)
	return screenRegion{region.Zoom, x, y, right - x, bottom - y}, right > x && bottom > y
}

func (v *documentView) scheduleRasterPatch(visible []sheet) bool {
	if len(v.rasterDamage) == 0 || v.proofMode() {
		return false
	}
	for _, target := range v.visibleDetails {
		if v.requestRasterPatch(target) {
			return true
		}
	}
	for _, page := range visible {
		if v.requestRasterPatch(screenTile{page: v.order[page.index]}) {
			return true
		}
	}
	return false
}

func (v *documentView) requestRasterPatch(target screenTile) bool {
	area, exists := v.rasterDamage[target]
	if !exists {
		return false
	}
	region, ok := v.patchRegion(target, area)
	if !ok {
		delete(v.rasterDamage, target)
		return false
	}
	patch := screenTile{target.page, region}
	req := v.request("screen")
	req.View.Page, req.View.Rotation = target.page, v.rotations[target.page-1]
	req.View.Region = &patch.screenRegion
	v.launchReader(req, target.page, 0, false, &patch, &target)
	return true
}

func (v *documentView) finishRasterPatch(job readerJob) {
	target, patch := *job.patch, *job.tile
	delete(v.rasterDamage, target)
	image := v.reading.frames[target.page].image
	if target.Zoom != 0 {
		image = v.detailFrames[target].image
	}
	if job.result.Status == "changed" {
		v.invalidSource = true
		v.app.reportStatus(job.result.Message)
		v.app.updateActions()
		v.invalidateCanvas()
	} else if !copyRasterPatch(image, job.frame.image, patch.X-target.X, patch.Y-target.Y) {
		if target.Zoom == 0 {
			v.reading.remove(target.page)
		} else if image != nil {
			v.detailBytes -= int64(len(image.pixels))
			delete(v.detailFrames, target)
		}
	} else if pagePixels, page := v.textSheetAt(0, 0, target.page); page != 0 {
		dirty := v.projectDetail(patch, pagePixels)
		v.invalidateRasterArea(rect{dirty.left - 2, dirty.top - 2, dirty.right + 2, dirty.bottom + 2})
	}
	v.settleContentPreview(target.page)
	v.schedule()
}

func copyRasterPatch(target, source *printBitmap, x, y int) bool {
	if target == nil || source == nil || x < 0 || y < 0 ||
		x+source.width > target.width || y+source.height > target.height {
		return false
	}
	stride := source.width * 4
	for row := range source.height {
		start := ((y+row)*target.width + x) * 4
		copy(target.pixels[start:start+stride], source.pixels[row*stride:(row+1)*stride])
	}
	return true
}
