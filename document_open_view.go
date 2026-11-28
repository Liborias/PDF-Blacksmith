package main

// Metadata alone (renderedPage=0), a secondary reader, another page, or
// a cancelled pending transition must not switch the user's tool.
func initialPageImageReady(pending, secondary, ready bool, order []int, current, renderedPage int) bool {
	return pending && !secondary && ready && current >= 0 && current < len(order) &&
		renderedPage > 0 && order[current] == renderedPage
}
