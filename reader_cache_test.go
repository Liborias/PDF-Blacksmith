package main

import "testing"

func TestReaderCacheLimit(t *testing.T) {
	const mib = 1 << 20
	for _, tc := range []struct{ available, cached, limit int64 }{
		{1024 * mib, 100 * mib, 0},
		{2048 * mib, 0, 1024 * mib / 10},
		{1948 * mib, 100 * mib, 1024 * mib / 10},
		{16 << 30, 0, maxReaderCache},
	} {
		if got := readerCacheLimit(tc.available, tc.cached); got != tc.limit {
			t.Errorf("readerCacheLimit(%d, %d) = %d, want %d", tc.available, tc.cached, got, tc.limit)
		}
	}
}
