package main

import (
	"runtime/debug"
	"unsafe"
)

const readerMemoryTimer = 1

var globalMemoryStatus = kernel32.NewProc("GlobalMemoryStatusEx")

type memoryStatus struct {
	length, load                                                   uint32
	totalPhysical, availablePhysical, totalCommit, availableCommit uint64
	totalVirtual, availableVirtual, availableExtended              uint64
}

func (v *documentView) cachedBytes() int64 {
	return v.reading.bytes + v.printPreview.bytes + v.detailBytes + v.canvas.bytes() + v.thumbnails.bytes
}

func (v *documentView) updateCacheLimit() {
	status := memoryStatus{length: uint32(unsafe.Sizeof(memoryStatus{}))}
	v.cacheLimit = 0
	if ok, _, _ := globalMemoryStatus.Call(uintptr(unsafe.Pointer(&status))); ok != 0 {
		cached := v.cachedBytes()
		comparing := v.app.comparison.tabs[0] != nil
		if comparing {
			other := v.app.comparison.view
			if v.secondary {
				other = v.app.view
			}
			cached += other.cachedBytes()
		}
		v.cacheLimit = readerCacheLimit(int64(min(status.availablePhysical, status.availableCommit)), cached)
		if comparing {
			v.cacheLimit /= 2
		}
	}
}

type pageCache struct {
	frames map[int]readerFrame
	failed map[int]bool
	bytes  int64
}

func (c *pageCache) clear() {
	c.frames = make(map[int]readerFrame)
	c.failed = make(map[int]bool)
	c.bytes = 0
}

func (c *pageCache) remove(number int) {
	if previous := c.frames[number].image; previous != nil {
		c.bytes -= int64(len(previous.pixels))
		delete(c.frames, number)
	}
	delete(c.failed, number)
}

func (c *pageCache) store(number int, frame readerFrame) {
	c.remove(number)
	c.frames[number] = frame
	c.bytes += int64(len(frame.image.pixels))
}

func (v *documentView) activeCache() *pageCache {
	if v.proofMode() {
		return &v.printPreview
	}
	return &v.reading
}

func (v *documentView) trimCache(cache *pageCache, budget int64, keepVisible bool) {
	if cache.bytes <= budget {
		return
	}
	_, height := v.viewportSize()
	visible := make(map[int]bool)
	if keepVisible {
		for _, item := range v.spread.visible(v.scrollY, height) {
			visible[v.order[item.index]] = true
		}
		visible[v.order[v.page]] = true
	}
	// Remove cached pages farthest from the currently displayed page.
	for left, right := 0, len(v.order)-1; left <= right && cache.bytes > budget; {
		index := right
		if v.page-left >= right-v.page {
			index = left
			left++
		} else {
			right--
		}
		if number := v.order[index]; !visible[number] && cache.frames[number].image != nil {
			cache.remove(number)
			if cache == &v.reading {
				delete(v.rasterDamage, screenTile{page: number})
			}
		}
	}
}

func (v *documentView) trimCaches() {
	if !v.ready() {
		return
	}
	v.trimThumbnails(v.thumbnailBudget())
	budget := max(0, v.cacheLimit-v.canvas.bytes()-v.thumbnails.bytes)
	v.trimDetails(min(detailCacheCeiling, budget))
	budget = max(0, budget-v.detailBytes)
	if v.proofMode() {
		v.trimCache(&v.printPreview, min(120<<20, budget), true)
		v.trimCache(&v.reading, max(0, budget-v.printPreview.bytes), false)
	} else {
		v.trimCache(&v.reading, budget, true)
		v.trimCache(&v.printPreview, min(120<<20, max(0, budget-v.reading.bytes)), false)
	}
}

func (v *documentView) refreshCache() {
	v.updateCacheLimit()
	previous := v.reading.bytes + v.printPreview.bytes + v.detailBytes + v.thumbnails.bytes
	v.trimCaches()
	released := previous - v.reading.bytes - v.printPreview.bytes - v.detailBytes - v.thumbnails.bytes
	if released > 0 {
		v.updateDocumentInfo()
		if released >= 32<<20 {
			go debug.FreeOSMemory()
		}
	}
	v.schedule()
}
