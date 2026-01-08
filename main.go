package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v3"
)

func renderHello(scr tcell.Screen) {
	w, h := scr.Size()
	for row := range h {
		if row == 0 || row == h-1 {
			for col := range w {
				scr.PutStr(col, row, "#")
			}
		} else {
			scr.PutStr(0, row, "#")
			scr.PutStr(w-1, row, "#")
		}
	}
	scr.PutStr(5, 5, "hello")
}

var keyInfoStrs = []string{"", ""}
var mouseInfoStrs = []string{"", ""}
var termInfoStrs = []string{"", ""}
var mx, my int
var drawChar = tcell.Key(0)
var drawStr = ""

func renderInfoStrs(scr tcell.Screen, yInfo *int, strs []string) {
	for i := range strs {
		scr.PutStr(5, *yInfo, strs[i])
		*yInfo++
	}
}

func renderBoard(scr tcell.Screen) {
	//if drawChar == tcell.KeyDel || drawChar == 0 || drawStr == ""{
	if drawStr == "" {
		scr.PutStr(mx, my, " ")
	} else {
		scr.PutStr(mx, my, drawStr)
	}
}

func draw(scr tcell.Screen) {
	scr.Clear()
	renderHello(scr)
	yInfo := 6
	renderInfoStrs(scr, &yInfo, keyInfoStrs)
	renderInfoStrs(scr, &yInfo, mouseInfoStrs)
	renderInfoStrs(scr, &yInfo, termInfoStrs)
	renderBoard(scr)
	scr.Show()
}

func main() {
	fmt.Println("vim-go")
	scr, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}
	err = scr.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}

	//scr.EnableMouse()
	//scr.EnableMouse(tcell.MouseMotionEvents | tcell.MouseButtonEvents)
	//scr.EnableMouse(tcell.MouseMotionEvents)
	scr.EnableMouse(tcell.MouseButtonEvents)
	//scr.DisablePaste()
	//scr.Clear()
	//renderHello(scr)
	//scr.Show()
	draw(scr)
	//scr.Beep()
	termInfoStrs[0], termInfoStrs[1] = scr.Terminal()
	draw(scr)
	scr.ShowCursor(20, 20)

	for {
		ev := <-scr.EventQ()

		switch ev := ev.(type) {
		case *tcell.EventKey:
			drawChar = ev.Key()
			drawStr = ev.Str()
			keyInfoStrs[0] = fmt.Sprintf("%04x", drawChar)
			keyInfoStrs[1] = fmt.Sprintf("%d %s", len(drawStr), drawStr)
			if drawChar == tcell.KeyEscape {
				scr.Fini()
				return
			}
		case *tcell.EventMouse:
			mouseInfoStrs[0] = fmt.Sprintf("%04x %04x", ev.Buttons(), ev.Modifiers())
			mx, my = ev.Position()
			mouseInfoStrs[1] = fmt.Sprintf("%d %d", mx, my)
		}

		draw(scr)
	}
}
