package main

type headerHistoryBox struct {
	left, top, right, bottom int
}

// Logical coordinates, before DPI scaling. Leave room after the brand and
// before document tabs; the same geometry is used for painting and hit testing.
func headerHistoryLayout(brandWidth, height int) (buttons [2]headerHistoryBox, tabsLeft int) {
	const size, gap, brandGap, tabGap = 27, 5, 8, 14
	top := (height - size) / 2
	for index := range buttons {
		left := brandWidth + brandGap + index*(size+gap)
		buttons[index] = headerHistoryBox{left, top, left + size, top + size}
	}
	return buttons, buttons[1].right + tabGap
}
