package main

import (
	"slices"
	"strings"
	"unsafe"
)

type documentFont struct {
	Name string
	Data []byte
}

func (a *app) clearInlineFonts() {
	p, e := &a.textTool, &a.textTool.inline
	if len(e.fontHandles) == 0 {
		return
	}
	e.loading = true
	setText(e.handle, "")
	for _, handle := range e.fontHandles {
		gdi32.NewProc("RemoveFontMemResourceEx").Call(handle)
	}
	e.fontHandles = nil
	p.fontNames = nil
	a.updateInlineFontChoices()
	e.loading = false
}

func (a *app) loadInlineFonts(frame *inlineTextFrame) {
	p := &a.textTool
	p.fontNames = map[string]string{}
	if p.inline.fontHandles == nil {
		p.inline.fontHandles = map[string]uintptr{}
	}
	for _, glyph := range frame.flow.glyphs() {
		key := p.model.Objects[glyph.run].Face
		if key == "" || p.fontNames[key] != "" {
			continue
		}
		font := p.model.Fonts[key]
		if len(font.Data) == 0 || len(font.Data) > 1<<20 {
			continue
		}
		if p.inline.fontHandles[key] == 0 {
			var loaded uint32
			handle, _, _ := gdi32.NewProc("AddFontMemResourceEx").Call(uintptr(unsafe.Pointer(&font.Data[0])),
				uintptr(len(font.Data)), 0, uintptr(unsafe.Pointer(&loaded)))
			p.inline.fontHandles[key] = handle
		}
		p.fontNames[key] = font.Name
	}
	a.updateInlineFontChoices()
}

func (a *app) updateInlineFontChoices() {
	p := &a.textTool
	keys := make([]string, 0, len(p.fontNames))
	for key := range p.fontNames {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, func(left, right string) int {
		if compare := strings.Compare(p.fontNames[left], p.fontNames[right]); compare != 0 {
			return compare
		}
		return strings.Compare(left, right)
	})
	p.fontKeys = append(keys, p.fonts...)
	labels := slices.Clone(p.fontKeys)
	for index := range keys {
		labels[index] = p.fontNames[keys[index]] + " (z PDF)"
	}
	choice := a.choices[p.font]
	choice.names, choice.index, choice.placeholder = labels, 0, ""
	setText(p.font, labels[0])
	invalidate(p.font)
}

func (a *app) inlineFontForRun(source textRun) string {
	p := &a.textTool
	if _, loaded := p.fontNames[source.Face]; loaded {
		return source.Face
	}
	name := source.Font
	if len(name) > 7 && name[6] == '+' {
		name = name[7:]
	}
	for _, system := range p.fonts {
		if strings.EqualFold(name, system) {
			return system
		}
	}
	return p.fonts[min(max(source.Family, 0), len(bundledTextFonts)-1)]
}
