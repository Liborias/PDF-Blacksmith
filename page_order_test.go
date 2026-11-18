package main

import (
	"slices"
	"testing"
)

func TestPlacePages(t *testing.T) {
	order := []int{8, 3, 7, 1, 4}
	tests := []struct {
		name     string
		selected []int
		position int
		want     []int
	}{
		{"one page moves to the start", []int{3}, 0, []int{1, 8, 3, 7, 4}},
		{"disjoint selection at the end", []int{0, 2}, 3, []int{3, 1, 4, 8, 7}},
		{"disjoint selection in the middle", []int{0, 4}, 1, []int{3, 8, 4, 7, 1}},
		{"same position", []int{2, 3}, 2, nil},
		{"position leaves too little room", []int{2, 3}, 4, nil},
		{"duplicate selection", []int{2, 2}, 1, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := placePages(order, test.selected, test.position); !slices.Equal(got, test.want) {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
	if !slices.Equal(order, []int{8, 3, 7, 1, 4}) {
		t.Fatal("the original order changed")
	}
}

func TestDropPosition(t *testing.T) {
	selected := []int{1, 3}
	for boundary, want := range []int{0, 1, 1, 2, 2, 3} {
		if got := dropPosition(selected, boundary); got != want {
			t.Errorf("boundary %d becomes %d, want %d", boundary, got, want)
		}
	}
}
