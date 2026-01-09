package main

import (
	"fmt"
	"math"
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
var mouseX, mouseY int
var charToDraw = tcell.Key(0)
var strToDraw = ""
var lastMouseX, lastMouseY = 0, 0
var running = true

type TBoardEvent struct {
	evType int
	dt     int64
}

var tEvTimeout = int64(0)
var tbEvCh = make(chan TBoardEvent)
var wg = sync.WaitGroup{}

type Vec2 struct {
	x, y float32
}

type Line struct {
	x1, y1, x2, y2 int
}

func (l *Line) render(scr tcell.Screen) {
	fx, fy := float32(l.x1), float32(l.y1)
	mx := float32(l.x2 - l.x1)
	my := float32(l.y2 - l.y1)
	maxM := mx
	if my > mx {
		maxM = my
	}
	mx = mx / maxM
	my = my / maxM
	scr.PutStr(int(fx), int(fy), "o")
	for {
		dx := float32(l.x2) - fx
		dy := float32(l.y2) - fy
		if dx == 0 && dy == 0 {
			break
		}
		d := math.Sqrt(float64(dx*dx + dy*dy))
		if d <= 0.1 {
			break
		}
		fx += mx
		fy += my
		scr.PutStr(int(fx), int(fy), "o")

	}
}

func renderInfoStrs(scr tcell.Screen, yInfo *int, strs []string) {
	for i := range strs {
		scr.PutStr(5, *yInfo, strs[i])
		*yInfo++
	}
}

var line = Line{
	x1: 16, y1: 2, x2: 23, y2: 12,
}

func renderBoard(scr tcell.Screen) {
	//if charToDraw == tcell.KeyDel || charToDraw == 0 || strToDraw == ""{
	if strToDraw == "" {
		scr.PutStr(lastMouseX, lastMouseY, " ")
	} else {
		scr.PutStr(lastMouseX, lastMouseY, strToDraw)
	}
	line.render(scr)
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

		tbEv := TBoardEvent{
			evType: 1,
		}

		switch ev := ev.(type) {
		case *tcell.EventKey:
			charToDraw = ev.Key()
			strToDraw = ev.Str()
			keyInfoStrs[0] = fmt.Sprintf("%04x", charToDraw)
			keyInfoStrs[1] = fmt.Sprintf("%d %s", len(strToDraw), strToDraw)
			if charToDraw == tcell.KeyEscape {
				tbEvCh <- TBoardEvent{
					evType: 2,
				}
				running = false
				wg.Done()
				return
			}
			tbEvCh <- tbEv
		case *tcell.EventMouse:
			mouseInfoStrs[0] = fmt.Sprintf("%04x %04x", ev.Buttons(), ev.Modifiers())
			mouseX, mouseY = ev.Position()
			lastMouseX, lastMouseY = mouseX, mouseY
			mouseInfoStrs[1] = fmt.Sprintf("%d %d", mouseX, mouseY)
			tbEvCh <- tbEv
		}

	}
}

func resetStrInfo() {
	//charToDraw = tcell.Key(0)
	//strToDraw = ""
	keyInfoStrs[0] = fmt.Sprintf("%04x", charToDraw)
	keyInfoStrs[1] = fmt.Sprintf("%d %s", len(strToDraw), strToDraw)
	mouseInfoStrs[0] = fmt.Sprintf("%04x %04x", 0, 0)
	mouseX, mouseY = 0, 0
	mouseInfoStrs[1] = fmt.Sprintf("%d %d", mouseX, mouseY)
}

func drawNotifier() {
	now := time.Now().UnixMilli()
	prev := now
	drawRemain := int64(25)
	prevDraw := now
	tbEv := TBoardEvent{
		evType: 0,
		dt:     0,
	}
	tbEvCh <- tbEv
	for running {
		now = time.Now().UnixMilli()
		dt := now - prev
		drawRemain -= dt
		if drawRemain <= 0 {
			tbEv.dt = now - prevDraw
			tbEvCh <- tbEv
			prevDraw = now
			drawRemain %= 25
			drawRemain += 25
		}
		time.Sleep(1 * time.Millisecond)
	}
	wg.Done()
}

func drawListener(scr tcell.Screen) {
	for tbEv := range tbEvCh {
		switch tbEv.evType {
		case 0:
			if tEvTimeout > 0 {
				tEvTimeout -= tbEv.dt
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

	scr.EnableMouse()
	//scr.EnableMouse(tcell.MouseMotionEvents | tcell.MouseButtonEvents)
	//scr.EnableMouse(tcell.MouseMotionEvents)
	//scr.EnableMouse(tcell.MouseButtonEvents)
	//scr.DisablePaste()

	resetStrInfo()
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
