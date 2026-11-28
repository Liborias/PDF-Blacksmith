package main

// Width allocation starts to the right of the existing reader controls.
// Preserve both legacy message fields before adding supplementary metadata.
type footerSlotWidths struct {
	mode, dirty, tool, help, status int
}

func fitFooterSlots(available, gap, mode, dirty, tool, helpMin, statusMin, statusPreferred int) footerSlotWidths {
	var p footerSlotWidths
	remaining := max(0, available)
	allocate := func(w int) int {
		if w <= 0 || remaining < w+gap {
			return 0
		}
		remaining -= w+gap
		return w
	}
	// The message field is mandatory whenever there is any room.
	p.status = min(statusMin, remaining)
	remaining -= p.status
	p.help = allocate(helpMin)
	// Dirty status is more important than decorative tool and view labels.
	p.dirty = allocate(dirty)
	p.mode = allocate(mode)
	p.tool = allocate(tool)
	extra := min(max(0, statusPreferred-p.status), remaining)
	p.status += extra
	remaining -= extra
	if p.help > 0 {
		p.help += remaining
	} else {
		p.status += remaining
	}
	return p
}
