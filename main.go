package main

import (
	"fmt"
	"os"
	"sync"
	"time"

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
var drawX, drawY = 0, 0
var running = true

type MyEv struct {
	evType int
	dt     int64
}

var tEvTimeout = int64(0)
var myEvCh = make(chan MyEv)
var wg = sync.WaitGroup{}

func renderInfoStrs(scr tcell.Screen, yInfo *int, strs []string) {
	for i := range strs {
		scr.PutStr(5, *yInfo, strs[i])
		*yInfo++
	}
}

func renderBoard(scr tcell.Screen) {
	//if drawChar == tcell.KeyDel || drawChar == 0 || drawStr == ""{
	if drawStr == "" {
		scr.PutStr(drawX, drawY, " ")
	} else {
		scr.PutStr(drawX, drawY, drawStr)
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

func handleEv(scr tcell.Screen) {
	for {
		ev := <-scr.EventQ()

		myEv := MyEv{
			evType: 1,
		}

		switch ev := ev.(type) {
		case *tcell.EventKey:
			drawChar = ev.Key()
			drawStr = ev.Str()
			keyInfoStrs[0] = fmt.Sprintf("%04x", drawChar)
			keyInfoStrs[1] = fmt.Sprintf("%d %s", len(drawStr), drawStr)
			if drawChar == tcell.KeyEscape {
				myEvCh <- MyEv{
					evType: 2,
				}
				running = false
				wg.Done()
				return
			}
			myEvCh <- myEv
		case *tcell.EventMouse:
			mouseInfoStrs[0] = fmt.Sprintf("%04x %04x", ev.Buttons(), ev.Modifiers())
			mx, my = ev.Position()
			drawX, drawY = mx, my
			mouseInfoStrs[1] = fmt.Sprintf("%d %d", mx, my)
			myEvCh <- myEv
		}

	}
}

func resetStrInfo() {
	//drawChar = tcell.Key(0)
	//drawStr = ""
	//keyInfoStrs[0] = fmt.Sprintf("%04x", drawChar)
	//keyInfoStrs[1] = fmt.Sprintf("%d %s", len(drawStr), drawStr)
	mouseInfoStrs[0] = fmt.Sprintf("%04x %04x", 0, 0)
	mx, my = 0, 0
	mouseInfoStrs[1] = fmt.Sprintf("%d %d", mx, my)
}

func drawNotifier() {
	now := time.Now().UnixMilli()
	prev := now
	drawRemain := int64(25)
	prevDraw := now
	myEv := MyEv{
		evType: 0,
		dt:     0,
	}
	myEvCh <- myEv
	for running {
		now = time.Now().UnixMilli()
		dt := now - prev
		drawRemain -= dt
		if drawRemain <= 0 {
			myEv.dt = now - prevDraw
			myEvCh <- myEv
			prevDraw = now
			drawRemain %= 25
			drawRemain += 25
		}
		time.Sleep(1 * time.Millisecond)
	}
	wg.Done()
}

func drawListener(scr tcell.Screen) {
	for myEv := range myEvCh {
		switch myEv.evType {
		case 0:
			if tEvTimeout > 0 {
				tEvTimeout -= myEv.dt
				if tEvTimeout <= 0 {
					resetStrInfo()
				}
			}
			draw(scr)
		case 1:
			tEvTimeout = 500
		case 2:
			wg.Done()
			return
		}
	}
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

	wg.Add(1)
	go handleEv(scr)

	wg.Add(1)
	go drawNotifier()

	wg.Add(1)
	go drawListener(scr)

	wg.Wait()
	scr.Fini()
}
