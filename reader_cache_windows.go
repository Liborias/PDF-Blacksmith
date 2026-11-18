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

func (v *documentView) updateCacheLimit() {
	status := memoryStatus{length: uint32(unsafe.Sizeof(memoryStatus{}))}
	v.cacheLimit = 0
	if ok, _, _ := globalMemoryStatus.Call(uintptr(unsafe.Pointer(&status))); ok != 0 {
		v.cacheLimit = readerCacheLimit(int64(min(status.availablePhysical, status.availableCommit)),
			v.reading.bytes+v.printPreview.bytes)
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
		}
	}
}

func (v *documentView) trimCaches() {
	if !v.ready() {
		return
	}
	if v.proofMode() {
		v.trimCache(&v.printPreview, min(120<<20, v.cacheLimit), true)
		v.trimCache(&v.reading, max(0, v.cacheLimit-v.printPreview.bytes), false)
	} else {
		v.trimCache(&v.reading, v.cacheLimit, true)
		v.trimCache(&v.printPreview, min(120<<20, max(0, v.cacheLimit-v.reading.bytes)), false)
	}
}

func (v *documentView) refreshCache() {
	v.updateCacheLimit()
	previous := v.reading.bytes + v.printPreview.bytes
	v.trimCaches()
	released := previous - v.reading.bytes - v.printPreview.bytes
	if released > 0 {
		v.updateDocumentInfo()
		if released >= 32<<20 {
			go debug.FreeOSMemory()
		}
	}
	v.schedule()
}
