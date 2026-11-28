package main

// Two consecutive outside samples close the hover preview; re-entry resets
// the grace interval without changing focus or invoking a command.
func fileMenuOutsideTicks(previous int, inside bool) int {
	if inside { return 0 }
	return min(2,max(0,previous)+1)
}
