package main

const (
	previewReserve = 1 << 30
	maxReaderCache = 768 << 20
)

func readerCacheLimit(available, cached int64) int64 {
	if available <= previewReserve {
		return 0
	}
	// Include this cache in the reclaimable pool so filling it does not
	// immediately lower its own budget.
	return min(maxReaderCache, (available-previewReserve+cached)/10)
}
