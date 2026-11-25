package main

import (
	"strconv"
	"strings"
)

type exportPagesError struct{ localizedMessage }

func (e exportPagesError) Error() string {
	return uiCzech.formatMessage(e.localizedMessage)
}

func parseExportPages(input string, pageCount int) ([]int, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, exportPagesError{localizedMessage{key: uiEmptyPageRange}}
	}
	input = strings.NewReplacer("–", "-", "—", "-").Replace(input)
	selected := make([]bool, pageCount)
	for part := range strings.SplitSeq(input, ",") {
		firstText, lastText, isRange := strings.Cut(part, "-")
		if !isRange {
			lastText = firstText
		}
		first, firstErr := strconv.Atoi(strings.TrimSpace(firstText))
		last, lastErr := strconv.Atoi(strings.TrimSpace(lastText))
		if firstErr != nil || lastErr != nil || first < 1 || last < first || last > pageCount {
			return nil, exportPagesError{localizedMessage{key: uiInvalidPageRange, args: []any{strings.TrimSpace(part), pageCount}}}
		}
		for index := first - 1; index < last; index++ {
			selected[index] = true
		}
	}
	var pages []int
	for index, included := range selected {
		if included {
			pages = append(pages, index+1)
		}
	}
	return pages, nil
}

func pageNumbers(positions []int) []int {
	numbers := make([]int, len(positions))
	for i, position := range positions {
		numbers[i] = position + 1
	}
	return numbers
}

func contiguousPages(sorted []int) bool {
	return len(sorted) > 1 && sorted[len(sorted)-1]-sorted[0] == len(sorted)-1
}
