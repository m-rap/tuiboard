package main

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"github.com/gdamore/tcell/v3"
)

func renderBorder(scr tcell.Screen) {
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
	//scr.PutStr(5, 5, "hello")
}

var keyInfoStrs = []string{"", ""}
var mouseInfoStrs = []string{"", ""}
var termInfoStrs = []string{"", ""}
var mouseX, mouseY int
var charToDraw = tcell.Key(0)
var strToDraw = ""
var lastMouseX, lastMouseY = 0, 0
var mouseEvCount = 0
var mouseEvHist = []*tcell.EventMouse{}
var running = true

const (
	TbEvExit int = iota + 0
	TbEvDraw
	TbEvKey
	TbEvMouse
)

type TBoardEvent struct {
	evType int
	dt     int64
	tcEv   tcell.Event
}

var tEvTimeout = int64(0)
var tbEvCh = make(chan TBoardEvent)
var wg = sync.WaitGroup{}

const (
	EditLineStart int = iota + 0
	EditLineEnd
)

var editState = EditLineStart

var lines = []*Line{}
var lineToAdd *Line = nil

type Vec2 struct {
	x, y float32
}

type Line struct {
	x1, y1, x2, y2 int
	c              string
}

func (l *Line) render(scr tcell.Screen) {
	fx, fy := float32(l.x1), float32(l.y1)
	mx := float32(l.x2 - l.x1)
	my := float32(l.y2 - l.y1)
	eucM := float32(math.Sqrt(float64(mx*mx + my*my)))
	mx = mx / eucM
	my = my / eucM
	//maxM := mx
	//if my > mx {
	//	maxM = my
	//}
	//mx = mx / maxM
	//my = my / maxM
	scr.PutStr(int(fx), int(fy), l.c)
	for i := 0; i < 1000; i++ {
		dx := float32(l.x2) - fx
		dy := float32(l.y2) - fy
		if dx == 0 && dy == 0 {
			break
		}
		d := math.Sqrt(float64(dx*dx + dy*dy))
		if d <= 1.5 {
			break
		}
		fx += mx
		fy += my
		scr.PutStr(int(fx), int(fy), l.c)

	}
}

func renderInfoStrs(scr tcell.Screen, yInfo *int, strs []string) {
	for i := range strs {
		scr.PutStr(5, *yInfo, strs[i])
		*yInfo++
	}
}

var line = Line{
	x1: 20, y1: 2,
	x2: 28, y2: 12,
	c: "o",
}

func renderBoard(scr tcell.Screen) {
	////if charToDraw == tcell.KeyDel || charToDraw == 0 || strToDraw == ""{
	//if strToDraw == "" {
	//	scr.PutStr(lastMouseX, lastMouseY, " ")
	//} else {
	//	scr.PutStr(lastMouseX, lastMouseY, strToDraw)
	//}
	//line.render(scr)
	scr.PutStr(2, 2, strToDraw)
	if editState == EditLineEnd {
		scr.PutStr(3, 2, "*")
	}
	for _, l := range lines {
		l.render(scr)
	}
	if lineToAdd != nil {
		scr.PutStr(lineToAdd.x1, lineToAdd.y1, lineToAdd.c)
	}
}

func draw(scr tcell.Screen) {
	scr.Clear()
	renderBorder(scr)
	//yInfo := 6
	//renderInfoStrs(scr, &yInfo, keyInfoStrs)
	//renderInfoStrs(scr, &yInfo, mouseInfoStrs)
	//renderInfoStrs(scr, &yInfo, termInfoStrs)

	//tmp := fmt.Sprintf("lines %d", len(lines))
	//scr.PutStr(5, yInfo, tmp)
	//yInfo++
	//if lineToAdd != nil {
	//	tmp := fmt.Sprintf("%d %d %d %d %s", lineToAdd.x1, lineToAdd.y1, lineToAdd.x2, lineToAdd.y2, lineToAdd.c)
	//	scr.PutStr(5, yInfo, tmp)
	//	yInfo++
	//}

	renderBoard(scr)
	scr.Show()
}

func handleKey(ev *tcell.EventKey) {
	charToDraw = ev.Key()
	strToDraw = ev.Str()
	keyInfoStrs[0] = fmt.Sprintf("%04x", charToDraw)
	keyInfoStrs[1] = fmt.Sprintf("%d %s", len(strToDraw), strToDraw)
}

func handleMouse(ev *tcell.EventMouse) {
	mouseEvHist = append(mouseEvHist, ev)
	mouseEvCount++
	mouseInfoStrs[0] = fmt.Sprintf("%04x %04x", ev.Buttons(), ev.Modifiers())
	mouseX, mouseY = ev.Position()
	lastMouseX, lastMouseY = mouseX, mouseY
	mouseInfoStrs[1] = fmt.Sprintf("%d %d", mouseX, mouseY)

	for i := len(mouseInfoStrs); i < 7; i++ {
		mouseInfoStrs = append(mouseInfoStrs, "")
	}

	mouseInfoStrs[2] = fmt.Sprintf("mouse evcount %d", mouseEvCount)
	strI := 3
	histI := len(mouseEvHist) - 1
	for strI < 6 && histI >= 0 {
		tmpBtns := mouseEvHist[histI].Buttons()
		tmpX, tmpY := mouseEvHist[histI].Position()
		mouseInfoStrs[strI] = fmt.Sprintf("%04x %d %d", tmpBtns, tmpX, tmpY)
		strI++
		histI--
	}
	if ev.Buttons()&tcell.Button1 > 0 {
		switch editState {
		case EditLineStart:
			if mouseX < 5 && mouseY < 5 {
				lines = []*Line{}
				break
			}
			var tmpStr string
			if strToDraw == "" {
				tmpStr = "o"
			} else {
				tmpStr = strToDraw
			}
			lineToAdd = &Line{
				x1: mouseX,
				y1: mouseY,
				c:  tmpStr,
			}
			editState = EditLineEnd
		case EditLineEnd:
			lineToAdd.x2 = mouseX
			lineToAdd.y2 = mouseY
			lines = append(lines, lineToAdd)
			lineToAdd = nil
			editState = EditLineStart
		}
	}
}

func handleEv(scr tcell.Screen) {
	for {
		ev := <-scr.EventQ()

		tbEv := TBoardEvent{}

		switch ev := ev.(type) {
		case *tcell.EventKey:
			key := ev.Key()
			if key == tcell.KeyEscape {
				running = false
				tbEv.evType = TbEvExit
				tbEvCh <- tbEv
				fmt.Println("wg.Done()")
				wg.Done()
				return
			}
			tbEv.evType = TbEvKey
			tbEv.tcEv = ev
			tbEvCh <- tbEv
		case *tcell.EventMouse:
			tbEv.evType = TbEvMouse
			tbEv.tcEv = ev
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

func drawNotifier(scr tcell.Screen) {
	now := time.Now().UnixMilli()
	prev := now
	drawRemain := int64(25)
	prevDraw := now
	tbEv := TBoardEvent{
		evType: TbEvDraw,
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
	scr.Fini()
	fmt.Println("wg.Done()")
	wg.Done()
}

func drawListener(scr tcell.Screen) {
	for tbEv := range tbEvCh {
		switch tbEv.evType {
		case TbEvDraw:
			if tEvTimeout > 0 {
				tEvTimeout -= tbEv.dt
				if tEvTimeout <= 0 {
					resetStrInfo()
				}
			}
			draw(scr)
		case TbEvKey:
			tEvTimeout = 500
			tcEv := tbEv.tcEv.(*tcell.EventKey)
			handleKey(tcEv)
		case TbEvMouse:
			tEvTimeout = 500
			tcEv := tbEv.tcEv.(*tcell.EventMouse)
			handleMouse(tcEv)
		case TbEvExit:
			fmt.Println("wg.Done()")
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
	//scr.ShowCursor(20, 20)

	wg.Add(1)
	go handleEv(scr)

	wg.Add(1)
	go drawNotifier(scr)

	wg.Add(1)
	go drawListener(scr)

	wg.Wait()
}
