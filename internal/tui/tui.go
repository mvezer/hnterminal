package tui

import (
	"fmt"
	"hnterminal/internal/config"
	"hnterminal/internal/utils"
	"slices"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
	"log"
	"os"
	"sync"
)

type Layout int

const (
	LayoutHorizontalGrid Layout = iota
	LayoutVerticalGrid
	LayoutFill
	LayoutFloat
)

// ----------------- TUI -----------------
type TUI struct {
	screen       tcell.Screen
	config       *config.Config
	root         *Component
	defaultStyle tcell.Style
	maxId        int
	drawMap      map[int]*Component
	mutex        sync.Mutex
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
		root:         nil,
		defaultStyle: defaultStyle,
		maxId:        -1,
		drawMap:      make(map[int]*Component),
	}
	screen.SetStyle(tui.defaultStyle)
	screen.EnableMouse()
	screen.EnablePaste()
	screen.Clear()

	// boxStyle := tcell.StyleDefault.Foreground(color.Green).Background(color.Purple)
	// newBoxStyle := tcell.StyleDefault.Foreground(color.Green).Background(color.Green)
	root := tui.NewBox()
	root.SetLayout(LayoutFill)
	tui.root = &root

	// set up logging
	f, err := os.OpenFile("hnreader.log", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(f)
	log.Println("Starting TUI")

	return tui
}

func (t *TUI) CalculateGridGeometry(component *Component) []int {
	components := component.GetSiblings()
	parentComponent := component.Parent()
	sizes := make([]int, len(components))
	targetSize := parentComponent.Width() - parentComponent.Padding()*2
	if parentComponent.Layout() == LayoutVerticalGrid {
		targetSize = parentComponent.Height() - parentComponent.Padding()*2
	}
	targetSize -= component.Padding() * 2
	defaultSize := targetSize / len(components)
	fixedSize := 0
	fixedComponents := make(map[int]bool, 0)
	calcutatedSize := 0
	for i, c := range components {
		max := c.MaxWidth()
		min := c.MaxHeight()
		if parentComponent.Layout() == LayoutVerticalGrid {
			max = c.MaxHeight()
			min = c.MinHeight()
		}
		if max > 0 && defaultSize > max {
			fixedComponents[i] = true
			sizes[i] = max
			fixedSize += max
		} else if min > 0 && defaultSize < min {
			fixedComponents[i] = true
			sizes[i] = min
			fixedSize += min
		} else {
			sizes[i] = defaultSize
		}
		calcutatedSize += sizes[i]
	}
	if len(fixedComponents) > 0 && calcutatedSize != targetSize {
		calcutatedSize = 0
		defaultSize = (targetSize - fixedSize) / (len(components) - len(fixedComponents))
		for i := range components {
			if !fixedComponents[i] {
				sizes[i] = defaultSize
			}
			calcutatedSize += sizes[i]
		}
	}
	if calcutatedSize != targetSize {
		for i := range components {
			if !fixedComponents[i] {
				sizes[i] += targetSize - calcutatedSize
				break
			}
		}
	}
	offset := parentComponent.Padding()
	ids := make([]int, len(components))
	for i, c := range components {
		if parentComponent.Layout() == LayoutHorizontalGrid {
			c.SetGeometry(offset, parentComponent.Padding(), sizes[i], parentComponent.Height()-parentComponent.padding*2)
		} else {
			c.SetGeometry(parentComponent.Padding(), offset, parentComponent.Width()-parentComponent.padding*2, sizes[i])
		}
		offset += sizes[i]
		ids[i] = c.Id()
	}
	return ids
}

// Traverse through the list of components and updates the geometry for the ones that have changed
// All the components are considered changed in the subtree of a changed component and also in the subtrees of the
// siblings of the changed component
func (t *TUI) UpdateGeometry(rootComponent *Component) {
	if rootComponent == nil {
		rootComponent = t.root
	}
	processed := make(map[int]bool) // the already processed components
	for _, c := range rootComponent.Traverse() {
		if processed[c.Id()] { // don't process the same component twice
			continue
		}
		screenWidth, screenHeight := t.screen.Size()
		if c.Floating() {
			c.SetGeometry(utils.Max(c.X(), 0), utils.Max(c.Y(), 0), utils.Min(c.Width(), screenWidth-c.X()), utils.Min(c.Height(), screenHeight-c.Y()))
		} else {
			if c.IsRoot() {
				c.SetGeometry(0, 0, screenWidth, screenHeight)
				processed[c.Id()] = true
			} else if c.Parent().Layout() == LayoutFill {
				c.SetGeometry(c.Parent().Padding(), c.Parent().Padding(), c.Parent().Width()-c.Parent().Padding(), c.Parent().Height()-c.Parent().Padding())
				processed[c.Id()] = true
			} else if !c.IsRoot() && (c.Parent().Layout() == LayoutHorizontalGrid || c.Parent().Layout() == LayoutVerticalGrid) {
				for _, i := range t.CalculateGridGeometry(c) {
					processed[i] = true
				}
			}
		}
	}
}

func (t *TUI) DrawAll() error {
	drawQueue := make([]*Component, 0)
	for _, c := range t.drawMap {
		drawQueue = append(drawQueue, c)
	}
	slices.SortFunc(drawQueue, func(a, b *Component) int {
		if a.Floating() && b.Floating() {
			return a.Id() - b.Id()
		} else if a.Floating() {
			return -1
		} else if b.Floating() {
			return 1
		}
		if a.ZIndex() == b.ZIndex() {
			return a.Id() - b.Id()
		}
		return a.ZIndex() - b.ZIndex()
	})
	for _, e := range drawQueue {
		e.Draw()
	}
	t.ClearDrawMap()
	return nil
}

func (t *TUI) AddToDrawMap(c *Component) {
	t.drawMap[c.Id()] = c
}

func (t *TUI) ClearDrawMap() {
	t.drawMap = make(map[int]*Component)
}

func (t *TUI) NextId() int {
	t.maxId++
	return t.maxId
}

func (t *TUI) Run() {
	box1 := t.NewBox()
	box1.SetStyle(tcell.StyleDefault.Foreground(color.White).Background(color.Green))
	box2 := t.NewBox()
	box2.SetLayout(LayoutVerticalGrid)
	box2.SetStyle(tcell.StyleDefault.Foreground(color.Red).Background(color.LightCyan))
	box3 := t.NewBox()
	box3.SetLayout(LayoutVerticalGrid)
	box3.SetStyle(tcell.StyleDefault.Foreground(color.White).Background(color.Red))
	for i := range 5 {
		b := t.NewBox()
		if i%2 == 0 {
			b.SetStyle(tcell.StyleDefault.Foreground(color.White).Background(color.Black))
		} else {
			b.SetStyle(tcell.StyleDefault.Foreground(color.Black).Background(color.White))
		}
		if i == 2 {
			t := t.NewText("Everyone is a genius. But if you judge a fish by its ability to climb a tree, it will live its whole life believing that it is stupid.")
			t.SetLayout(LayoutFill)
			b.AddChild(&t)
		}
		if i == 3 {
			b.SetMaxHeight(2)
			b.SetStyle(tcell.StyleDefault.Foreground(color.Black).Background(color.Yellow))
		}
		if i == 1 {
			b.SetMinHeight(25)
			b.SetStyle(tcell.StyleDefault.Foreground(color.Black).Background(color.Yellow))
		}
		if i == 4 {
			b.Spec().(*Box).SetBorder(BorderStyleDouble)
		}
		box2.AddChild(&b)
	}
	t.root.SetLayout(LayoutHorizontalGrid)
	t.root.SetPadding(1)
	t.root.AddChild(&box1)
	t.root.AddChild(&box2)
	t.root.AddChild(&box3)
	defer t.Quit()
	for {
		t.UpdateGeometry(nil)
		t.DrawAll()
		t.screen.Show()
		ev := <-t.screen.EventQ()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			t.screen.Sync()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyCtrlC, tcell.KeyEscape:
				t.Quit()
				return
			// case tcell.KeyEnter:
			// case tcell.KeyRune:
			// 	fmt.Printf("Rune: %c\n", ev.())
			default:
				log.Printf("Key: %s\n", ev.Name())
			}
		}
	}
}
func (t *TUI) Quit() {
	fmt.Println("Bye")
	maybePanic := recover()
	t.screen.Fini()
	if maybePanic != nil {
		panic(maybePanic)
	}
}
