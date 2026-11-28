package main

type toolNavigationRow struct { top,height int }
type toolNavigationPlan struct {
	heading int
	rows []toolNavigationRow
	separators []int
}

// Reserve decorative space only after every row can fit at its compact
// height. Preserve the existing preferred step, including row spacing.
func planToolNavigation(height int, groups []int, margin,gap,normalStep,compactStep,heading,separator int) toolNavigationPlan {
	p:=toolNavigationPlan{rows:make([]toolNavigationRow,len(groups))}
	if len(groups)==0 || height<=0 { return p }
	count:=0
	for i:=1;i<len(groups);i++ { if groups[i]!=groups[i-1] { count++ } }
	available:=max(0,height-2*margin)
	minimum:=len(groups)*compactStep
	if available>=minimum+heading { p.heading=heading; available-=heading }
	if available>=minimum+count*separator { available-=count*separator } else { separator=0 }
	step:=min(normalStep,available/len(groups))
	y:=margin+p.heading
	for i,group:=range groups {
		if i>0 && group!=groups[i-1] && separator>0 {
			p.separators=append(p.separators,y+separator/2)
			y+=separator
		}
		p.rows[i]=toolNavigationRow{y,max(0,step-min(gap,max(0,step-1)))}
		y+=step
	}
	return p
}
