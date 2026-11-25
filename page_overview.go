package main

import "slices"

type overviewLayout struct {
	count, columns, cardWidth, cardHeight, gap, left, height int
}

func layoutOverview(count, width, preferred, gap, caption int) overviewLayout {
	if count <= 0 || width <= 0 {
		return overviewLayout{}
	}
	cardWidth := min(preferred, max(1, width-2*gap))
	columns := min(count, max(1, (width-gap)/(cardWidth+gap)))
	cardHeight := cardWidth*4/3 + caption
	rows := (count-1)/columns + 1
	return overviewLayout{
		count: count, columns: columns, cardWidth: cardWidth, cardHeight: cardHeight, gap: gap,
		left:   (width - columns*cardWidth - (columns-1)*gap) / 2,
		height: rows*(cardHeight+gap) + gap,
	}
}

func (grid overviewLayout) card(index int) (int, int) {
	return grid.left + index%grid.columns*(grid.cardWidth+grid.gap),
		grid.gap + index/grid.columns*(grid.cardHeight+grid.gap)
}

func (grid overviewLayout) visible(top, height int) (int, int) {
	if grid.columns == 0 || height <= 0 {
		return 0, 0
	}
	stride := grid.cardHeight + grid.gap
	firstRow := max(0, top/stride)
	lastRow := max(firstRow, (top+height-1)/stride+1)
	return min(grid.count, firstRow*grid.columns), min(grid.count, lastRow*grid.columns)
}

func (grid overviewLayout) at(x, y int) int {
	if grid.columns == 0 || x < grid.left || y < grid.gap {
		return -1
	}
	x, y = x-grid.left, y-grid.gap
	column, row := x/(grid.cardWidth+grid.gap), y/(grid.cardHeight+grid.gap)
	index := row*grid.columns + column
	if column >= grid.columns || index >= grid.count ||
		x%(grid.cardWidth+grid.gap) >= grid.cardWidth || y%(grid.cardHeight+grid.gap) >= grid.cardHeight {
		return -1
	}
	return index
}

func overviewSelection(selected []int, anchor, page int, toggle, extend bool) []int {
	if extend {
		first, last := min(anchor, page), max(anchor, page)
		rangeSelection := make([]int, last-first+1)
		for i := range rangeSelection {
			rangeSelection[i] = first + i
		}
		return rangeSelection
	}
	if !toggle {
		return []int{page}
	}
	picked := slices.Clone(selected)
	index, exists := slices.BinarySearch(picked, page)
	if exists {
		return slices.Delete(picked, index, index+1)
	}
	return slices.Insert(picked, index, page)
}
