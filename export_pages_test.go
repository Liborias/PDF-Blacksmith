package main

import (
	"slices"
	"testing"
)

func TestParseExportPages(t *testing.T) {
	got, err := parseExportPages(" 8, 3–5, 1, 4, 5-6 ", 9)
	if want := []int{1, 3, 4, 5, 6, 8}; err != nil || !slices.Equal(got, want) {
		t.Fatalf("parseExportPages got %v, %v; want %v", got, err, want)
	}
	for _, input := range []string{"", "0", "10", "3-2", "1-10", "2-", "1,,2", "2.5"} {
		if got, err := parseExportPages(input, 9); err == nil {
			t.Errorf("parseExportPages(%q) got %v, want error", input, got)
		}
	}
}
