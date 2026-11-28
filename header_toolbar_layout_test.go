package main

import "testing"

func TestHeaderToolbarPriorities(t *testing.T) {
	icons := []int{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17}
	optional := []int{1,2,3,4,5}
	emergency := []int{6,7,8,9,10,11,12,13,14,15,16,17}
	for _, dpi := range []int{96,120,144,192} {
		d := func(n int) int { return n*dpi/96 }
		for _, menu := range []int{340,410,480,550} {
			for width := 720; width <= 2200; width++ {
				available := d(width)-d(menu)-d(28)
				p := planHeaderToolbar(available, icons, optional, emergency,
					d(30),d(4),d(27),d(2),d(78),d(112),d(145),d(76))
				if p.used > max(0,available) {
					t.Fatalf("overflow at width=%d menu=%d dpi=%d: %+v",width,menu,dpi,p)
				}
				for i,id := range optional {
					if p.hidden[id] && i>0 && !p.hidden[optional[i-1]] {
						t.Fatalf("wrong hide order: %+v",p.hidden)
					}
				}
				if available >= len(icons)*(d(30)+d(4))+d(145)+d(76)+2*d(4) &&
					(len(p.hidden)!=0 || p.search!=d(145) || p.zoom!=d(76)) {
					t.Fatalf("wide layout failed to restore controls: %+v",p)
				}
				if available >= (len(icons)-len(optional))*(d(27)+d(2)) {
					for _,id := range emergency {
						if p.hidden[id] { t.Fatalf("unnecessary core hide: %d",id) }
					}
				}
			}
		}
	}
}

func TestHeaderPageNavigationState(t *testing.T) {
	for _, tc := range []struct {
		ready, busy bool
		page, total int
		want string
		previous, next bool
	}{
		{false, false, 0, 0, "— / —", false, false},
		{true, false, 0, 1, "1 / 1", false, false},
		{true, false, 0, 12, "1 / 12", false, true},
		{true, false, 5, 12, "6 / 12", true, true},
		{true, false, 11, 12, "12 / 12", true, false},
		{true, true, 5, 12, "6 / 12", false, false},
		{true, false, 12, 12, "— / —", false, false},
	} {
		s, p, n := headerPageNavigationState(tc.ready, tc.busy, tc.page, tc.total)
		if s != tc.want || p != tc.previous || n != tc.next {
			t.Fatalf("unexpected page state for %+v: %q %v %v",tc,s,p,n)
		}
	}
}

func TestHeaderPageNavigationFits(t *testing.T) {
	base := headerToolbarPlan{search:145,zoom:76,hidden:map[int]bool{1:true}}
	for _, tc := range []struct {
		candidate headerToolbarPlan
		want bool
	}{
		{headerToolbarPlan{search:112,zoom:76,hidden:map[int]bool{1:true}},true},
		{headerToolbarPlan{search:0,zoom:76,hidden:map[int]bool{1:true}},false},
		{headerToolbarPlan{search:145,zoom:0,hidden:map[int]bool{1:true}},false},
		{headerToolbarPlan{search:145,zoom:76,hidden:map[int]bool{1:true,2:true}},false},
	} {
		if got:=headerPageNavigationFits(base,tc.candidate,78); got!=tc.want {
			t.Fatalf("unexpected navigation fit for %+v: %v",tc,got)
		}
	}
}
