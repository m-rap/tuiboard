package main

import (
	"fmt"
	"unicode/utf8"

	"github.com/gdamore/tcell/v3"
)

type DoubleBufferScreen struct {
	Scr  tcell.Screen
	w, h int
	fb1  [][]rune
	fb2  [][]rune
	Fb   *[][]rune
}

func NewDoubleBufferScreen() (*DoubleBufferScreen, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	s.Init()
	ds := &DoubleBufferScreen{
		Scr: s,
		fb1: [][]rune{},
		fb2: [][]rune{},
		w:   0,
		h:   0,
	}
	ds.Fb = &ds.fb1
	return ds, nil
}

func (scr *DoubleBufferScreen) SwapBuffer() {
	if scr.Fb == &scr.fb1 {
		scr.Fb = &scr.fb2
	} else {
		scr.Fb = &scr.fb1
	}
}

func (scr *DoubleBufferScreen) Clear() {
	for row := range *scr.Fb {
		for col := range (*scr.Fb)[row] {
			(*scr.Fb)[row][col] = ' '
		}
	}
}

func (scr *DoubleBufferScreen) Draw() {
	for row := range *scr.Fb {
		scr.Scr.PutStr(0, row, string((*scr.Fb)[row]))
	}
}

func resizeFb(fb *[][]rune, w, h int) {
	for i := 0; i < h; i++ {
		if i >= len(*fb) {
			line := []rune{}
			for col := 0; col < w; col++ {
				line = append(line, ' ')
			}
			*fb = append(*fb, line)
		}
	}
}

func (scr *DoubleBufferScreen) CheckSize() (int, int) {
	scr.w, scr.h = scr.Scr.Size()
	resizeFb(&scr.fb1, scr.w, scr.h)
	resizeFb(&scr.fb2, scr.w, scr.h)
	return scr.w, scr.h
}

func (scr *DoubleBufferScreen) PutStr(x, y int, str string) {
	w, h := scr.w, scr.h
	if x < 0 || x >= w || y < 0 || y >= h {
		return
	}
	//logToFile(fmt.Sprintf("PutStr %d %d %s w h %d %d\n", x, y, str, w, h))
	stri := 0
	col := x
	for stri < len(str) && col < w {
		r, runeWidth := utf8.DecodeRuneInString(str[stri:])
		(*scr.Fb)[y][col] = r
		logToFile(fmt.Sprintf("r %c %d stri %d col %d\n", r, runeWidth, stri, col))
		stri += runeWidth
		col++
	}
}
