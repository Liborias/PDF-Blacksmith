package main

import "slices"

// placePages inserts the selected positions as a block at its final index.
// A nil result means the requested order was invalid or unchanged.
func placePages(order, selected []int, first int) []int {
	if len(selected) == 0 || first < 0 || first > len(order)-len(selected) {
		return nil
	}
	result := make([]int, len(order))
	next, picked := 0, 0
	for index, page := range order {
		if picked < len(selected) && index == selected[picked] {
			result[first+picked] = page
			picked++
			continue
		}
		if next == first {
			next += len(selected)
		}
		if next >= len(result) {
			return nil
		}
		result[next] = page
		next++
	}
	if picked != len(selected) || slices.Equal(order, result) {
		return nil
	}
	return result
}

// dropPosition converts an insertion boundary in the old list into the final index.
func dropPosition(selected []int, boundary int) int {
	before, _ := slices.BinarySearch(selected, boundary)
	return boundary - before
}
