package tui

import (
	"hnterminal/internal/config"
	"hnterminal/internal/utils"
	"log"

	"sync"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

// ----------------- TUI -----------------
type TUI struct {
	screen       tcell.Screen
	config       *config.Config
	defaultStyle tcell.Style
	maxId        int
	drawMap      map[int]*BaseComponent
	mutex        sync.Mutex
	root         *BaseComponent
}

func New(config *config.Config) *TUI {
	defaultStyle := tcell.StyleDefault.Background(color.Reset).Foreground(color.White)
	screen, err := tcell.NewScreen()
	if err != nil {
		utils.HandleError(err, utils.ErrorSeverityFatal)
	}
	if err := screen.Init(); err != nil {
		utils.HandleError(err, utils.ErrorSeverityFatal)
	}
	tui := &TUI{
		screen:       screen,
		config:       config,
		defaultStyle: defaultStyle,
		maxId:        -1,
		drawMap:      make(map[int]*BaseComponent),
	}
	screen.SetStyle(tui.defaultStyle)
	screen.EnableMouse()
	screen.EnablePaste()

	root := NewBox(HorizontalGrid)
	screenWidth, screenHeight := screen.Size()
	root.width = screenWidth
	root.height = screenHeight
	root.x = 0
	root.y = 0
	root.floating = true
	tui.root = &root
	tui.root.dirty = true

	utils.InitLogFile()

	return tui
}

var storiesList BaseComponent
var commentsList BaseComponent

func (t *TUI) Init() {
	t.root.SetStyle(tcell.StyleDefault.Foreground(color.White).Background(color.Purple))
	t.root.SetLayout(HorizontalGrid)
	storiesList = NewBox(FixedWidth)
	storiesList.SetStyle(tcell.StyleDefault.Foreground(color.White).Background(color.Green))
	storiesList.SetWidthPercent(40)
	t.root.AddChild(&storiesList)
	commentsList := NewBox(FixedWidth)
	commentsList.SetStyle(tcell.StyleDefault.Foreground(color.White).Background(color.Red))
	t.root.AddChild(&commentsList)

	story1 := NewBox(HorizontalGrid)
	story1.SetStyle(tcell.StyleDefault.Foreground(color.White).Background(color.Blue))
	storiesList.AddChild(&story1)

	story1Left := NewBox(FixedWidth)
	story1Left.SetWidthPercent(50)
	story1Left.SetPadding(Padding{4, 1, 5, 2})
	story1Left.SetStyle(tcell.StyleDefault.Foreground(color.Yellow).Background(color.Gray))
	story1LeftText := NewText("Story 1 with a long long text to test if the wrapping and the automatic height calculation works or not, so I can fix it if something is not looking alright", FixedWidth)
	story1LeftText.SetStyle(tcell.StyleDefault.Foreground(color.White).Background(color.Blue))
	story1LeftText.kind.(*Text).SetAlignment(TextAlignLeft)
	story1Left.AddChild(&story1LeftText)
	story1.AddChild(&story1Left)

	story1Right := NewBox(FixedWidth)
	story1Right.SetStyle(tcell.StyleDefault.Foreground(color.Black).Background(color.LightCyan))
	story1Right.kind.(*Box).SetBorderStyle(BorderStyleRounded)
	story1Right.kind.(*Box).SetBorder(Border{true, false, true, true})
	story1Right.SetPadding(Padding{3, 1, 3, 1})
	story1.AddChild(&story1Right)
	story1RightText := NewText("Just a short text, nothing to see here", FixedWidth)
	story1RightText.SetStyle(tcell.StyleDefault.Foreground(color.Pink).Background(color.BlueViolet))
	story1RightText.kind.(*Text).SetAlignment(TextAlignLeft)
	story1Right.AddChild(&story1RightText)
}

func (t *TUI) UpdateRoot() {
	w, h := t.screen.Size()
	t.root.width = w
	t.root.height = h
	t.root.x = 0
	t.root.y = 0
	t.root.kind.OnUpdate(t.root)
}

func (t *TUI) Run() {
	log.Println("Running")
	t.screen.Clear()
	t.screen.Show()
	defer t.Quit()
	for {
		t.Draw()
		ev := <-t.screen.EventQ()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			t.Draw()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyCtrlC, tcell.KeyEscape:
				t.Quit()
				return
			case tcell.KeyRight:
				storiesList.SetWidthPercent(min(storiesList.WidthPercent()+1, 100))
			case tcell.KeyLeft:
				storiesList.SetWidthPercent(max(storiesList.WidthPercent()-1, 5))
			// case tcell.KeyLeft:
			// 	box1.SetMinWidth(box1.MinWidth() - 1)
			// 	box1.SetMaxWidth(box1.MaxWidth() - 1)
			// 	t.UpdateGeometry(nil)
			// case tcell.KeyEnter:
			// 	toggleFloatBox(t)
			// case tcell.KeyEnter:
			// case tcell.KeyRune:
			// 	fmt.Printf("Rune: %c\n", ev.())
			default:
				log.Printf("Key: %s\n", ev.Name())
			}
		}
	}
}
func (t *TUI) Draw() {
	layoutStack := make([]*BaseComponent, 0)
	for root := range t.root.TraverseFloating() {
		isDirty := false
		parent := root.Find(func(c *BaseComponent) bool {
			return c.dirty
		})
		log.Printf("Parent is %s", parent.String())
		if parent == nil {
			continue
		}
		if parent.parent != nil {
			parent = parent.parent
			parent.dirty = true
		}
		if parent.IsRoot() {
			t.UpdateRoot()
		}
		for c := range parent.TraverseNonFloating() {
			if c.dirty && !isDirty {
				parent := c
				if c.parent != nil {
					parent = c.parent
				}
				isDirty = true
				for componentToReset := range parent.Traverse() {
					if !componentToReset.IsRoot() {
						componentToReset.ResetGeometry()
					}
				}
			}
			if isDirty {
				if LayoutFuncs[c.layout](c) {
					layoutStack = append(layoutStack, c)
				}
			}
		}
		for i := len(layoutStack) - 1; i >= 0; i-- {
			LayoutFuncs[layoutStack[i].layout](layoutStack[i])
		}
	}
	for root := range t.root.TraverseFloating() {
		for c := range root.TraverseNonFloating() {
			if c.dirty {
				// render only if the component is not fully covered and has a width and height
				if !((c.layout == VerticalGrid || c.layout == HorizontalGrid) && c.HasChildren()) && (c.height > 0 && c.width > 0) {
					c.Draw(t)
				}
			}
		}
	}
	t.screen.Sync()
}

func (t *TUI) Quit() {
	maybePanic := recover()
	t.screen.Fini()
	if maybePanic != nil {
		panic(maybePanic)
	}
}
