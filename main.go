package main

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"github.com/gdamore/tcell/v3"
)

var scr *DoubleBufferScreen

func drawBorder() {
	w, h := scr.CheckSize()
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
var charToDraw = tcell.KeyRune
var strToDraw = "#"
var lastMouseX, lastMouseY = 0, 0
var cursorX, cursorY = 0, 0
var mouseEvCount = 0
var mouseEvHist = []*tcell.EventMouse{}
var running = true
var drawMutex sync.Mutex
var logFile *os.File = nil

func logToFile(str string) {
	if logFile == nil {
		var err error
		logFile, err = os.Create("main.log")
		if err != nil {
			return
		}
	}
	logFile.WriteString(str)
}

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

// var tbEvCh = make(chan TBoardEvent)
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

func (l *Line) draw() {
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
		if d < 1.0 {
			break
		}
		fx += mx
		fy += my
		scr.PutStr(int(fx), int(fy), l.c)

	}
}

func drawInfoStrs(yInfo *int, strs []string) {
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

func drawBoard() {
	////if charToDraw == tcell.KeyDel || charToDraw == 0 || strToDraw == ""{
	//if strToDraw == "" {
	//	scr.PutStr(lastMouseX, lastMouseY, " ")
	//} else {
	//	scr.PutStr(lastMouseX, lastMouseY, strToDraw)
	//}
	//line.draw(scr)
	scr.PutStr(2, 2, "clear")
	scr.PutStr(2, 7, "apply")
	scr.PutStr(10, 2, "char: "+strToDraw)
	scr.PutStr(15, 2, strToDraw)
	if editState == EditLineEnd {
		scr.PutStr(10, 2, "char: "+strToDraw+" *")
	} else {
		scr.PutStr(10, 2, "char: "+strToDraw)
	}
	drawMutex.Lock()
	for _, l := range lines {
		l.draw()
	}
	if lineToAdd != nil {
		scr.PutStr(lineToAdd.x1, lineToAdd.y1, lineToAdd.c)
	}
	drawMutex.Unlock()
	scr.Scr.ShowCursor(cursorX, cursorY)
}

func drawIter() {
	scr.CheckSize()
	scr.Clear()
	//drawBorder()
	//yInfo := 6
	//drawInfoStrs(&yInfo, keyInfoStrs)
	//drawInfoStrs(&yInfo, mouseInfoStrs)
	//drawInfoStrs(&yInfo, termInfoStrs)

	//tmp := fmt.Sprintf("lines %d", len(lines))
	//scr.PutStr(5, yInfo, tmp)
	//yInfo++
	//if lineToAdd != nil {
	//	tmp := fmt.Sprintf("%d %d %d %d %s", lineToAdd.x1, lineToAdd.y1, lineToAdd.x2, lineToAdd.y2, lineToAdd.c)
	//	scr.PutStr(5, yInfo, tmp)
	//	yInfo++
	//}

	drawBoard()
	scr.Draw()
	scr.Scr.Show()
	scr.SwapBuffer()
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
		isDrawArea := false
		if mouseX < 10 {
			if mouseY < 5 {
				if editState == EditLineStart {
					lines = []*Line{}
				}
			} else if mouseY >= 5 && mouseY < 10 {
				switch editState {
				case EditLineStart:
					var tmpStr string
					if strToDraw == "" {
						tmpStr = "o"
					} else {
						tmpStr = strToDraw
					}
					drawMutex.Lock()
					lineToAdd = &Line{
						x1: cursorX,
						y1: cursorY,
						c:  tmpStr,
					}
					editState = EditLineEnd
					drawMutex.Unlock()
				case EditLineEnd:
					drawMutex.Lock()
					lineToAdd.x2 = cursorX
					lineToAdd.y2 = cursorY
					lines = append(lines, lineToAdd)
					lineToAdd = nil
					editState = EditLineStart
					drawMutex.Unlock()
				}

			} else {
				isDrawArea = true
			}
		} else {
			isDrawArea = true
		}
		if isDrawArea {
			cursorX = mouseX
			cursorY = mouseY
		}
	}
}

func handleEv() {
	for {
		ev := <-scr.Scr.EventQ()

		//tbEv := TBoardEvent{}

		switch ev := ev.(type) {
		case *tcell.EventKey:
			key := ev.Key()
			if key == tcell.KeyEscape {
				running = false
				//tbEv.evType = TbEvExit
				//tbEvCh <- tbEv
				fmt.Println("wg.Done()")
				wg.Done()
				return
			}
			//tbEv.evType = TbEvKey
			//tbEv.tcEv = ev
			//tbEvCh <- tbEv
			handleKey(ev)
		case *tcell.EventMouse:
			//tbEv.evType = TbEvMouse
			//tbEv.tcEv = ev
			//tbEvCh <- tbEv
			handleMouse(ev)
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
	//tbEv := TBoardEvent{
	//	evType: TbEvDraw,
	//	dt:     0,
	//}
	//tbEvCh <- tbEv
	for running {
		now = time.Now().UnixMilli()
		dt := now - prev
		drawRemain -= dt
		if drawRemain <= 0 {
			//tbEv.dt = now - prevDraw
			//tbEvCh <- tbEv
			dt := now - prevDraw
			if tEvTimeout > 0 {
				//tEvTimeout -= tbEv.dt
				tEvTimeout -= dt
				if tEvTimeout <= 0 {
					resetStrInfo()
				}
			}
			drawIter()
			prevDraw = now
			drawRemain %= 25
			drawRemain += 25
		}
		time.Sleep(1 * time.Millisecond)
	}
	scr.Scr.Fini()
	fmt.Println("wg.Done()")
	wg.Done()
}

func main() {
	fmt.Println("vim-go")
	var err error
	scr, err = NewDoubleBufferScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return
	}

	scr.Scr.EnableMouse()
	//scr.EnableMouse(tcell.MouseMotionEvents | tcell.MouseButtonEvents)
	//scr.EnableMouse(tcell.MouseMotionEvents)
	//scr.EnableMouse(tcell.MouseButtonEvents)
	//scr.DisablePaste()

	resetStrInfo()
	drawIter()
	//scr.Beep()
	termInfoStrs[0], termInfoStrs[1] = scr.Scr.Terminal()
	drawIter()
	//scr.ShowCursor(20, 20)

	wg.Add(1)
	go handleEv()

	wg.Add(1)
	go drawNotifier()

	wg.Wait()

	if logFile != nil {
		logFile.Close()
	}
}
