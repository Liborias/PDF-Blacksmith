package main

import "testing"

func TestHeaderHistoryLayout(t *testing.T) {
	for _, brandWidth := range []int{48, 200} {
		buttons, tabsLeft := headerHistoryLayout(brandWidth, 50)
		for _, dpi := range []int{96, 120, 144, 192} {
			scale := func(n int) int { return n * dpi / 96 }
			if scale(buttons[0].left) <= scale(brandWidth) ||
				scale(buttons[0].right) >= scale(buttons[1].left) ||
				scale(buttons[1].right) >= scale(tabsLeft) {
				t.Fatalf("overlapping brand, history or tabs at %d DPI", dpi)
			}
			for _, button := range buttons {
				if button.top < 6 || button.bottom > 49 ||
					button.right-button.left != 27 || button.bottom-button.top != 27 {
					t.Fatalf("invalid header button: %+v", button)
				}
			}
			// Even at a conservative 720 logical px, a minimum 84 px tab
			// and the 28 px tab list fit before the 108 px window controls.
			if scale(tabsLeft)+scale(84)+scale(28) > scale(720)-scale(108)-scale(4) {
				t.Fatalf("no room for document tab at %d DPI", dpi)
			}
		}
	}
}
