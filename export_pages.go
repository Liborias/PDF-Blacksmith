package main

import (
	"fmt"
	"strconv"
	"strings"
)

func parseExportPages(input string, pageCount int) ([]int, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("Zadejte výběr stran, například 1, 3–5, 8.")
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
			return nil, fmt.Errorf("Neplatná strana nebo rozsah %q. Zadejte strany od 1 do %d.", strings.TrimSpace(part), pageCount)
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
