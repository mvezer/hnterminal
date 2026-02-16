package tui

import (
	"fmt"
	"hnterminal/internal/config"
	"hnterminal/internal/utils"
	"slices"
	"strings"

	"log"
	"os"
	"sync"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

type Layout int

const (
	LayoutHorizontalGrid Layout = iota
	LayoutVerticalGrid
	LayoutFill
	LayoutFloat
)

type HorizontalAlignment int

const (
	HorizontalAlignmentLeft HorizontalAlignment = iota
	HorizontalAlignmentCenter
	HorizontalAlignmentRight
)

type VerticalAlignment int

const (
	VerticalAlignmentTop VerticalAlignment = iota
	VerticalAlignmentCenter
	VerticalAlignmentBottom
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
	// Remove floating components
	for i := range components {
		if components[i].Floating() {
			components = components[:i]
			break
		}
	}
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
		min := c.MinWidth()
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
			x := 0
			y := 0
			switch c.HorizontalAlignment() {
			case HorizontalAlignmentLeft:
				x = 0
			case HorizontalAlignmentCenter:
				x = (screenWidth - c.Width()) / 2
			case HorizontalAlignmentRight:
				x = screenWidth - c.Width()
			}
			switch c.VerticalAlignment() {
			case VerticalAlignmentTop:
				y = 0
			case VerticalAlignmentCenter:
				y = (screenHeight - c.Height()) / 2
			case VerticalAlignmentBottom:
				y = screenHeight - c.Height()
			}
			c.SetGeometry(x, y, c.Width(), c.Height())
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
		// check if the component is fully covered or not
		if !c.FullyCovered() {
			// if not fully covered we add it to the draw queue
			drawQueue = append(drawQueue, c)
		}
	}
	// sort the draw queue by z-index
	slices.SortFunc(drawQueue, func(a, b *Component) int {
		if a.ZIndex() == b.ZIndex() {
			return a.Id() - b.Id()
		}
		return a.ZIndex() - b.ZIndex()
	})
	var drawLog strings.Builder
	for _, e := range drawQueue {
		e.Draw()
		fmt.Fprintf(&drawLog, "[%d]", e.Id())
	}
	t.ClearDrawMap()
	log.Println(drawLog.String())
	t.screen.Show()
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

func (t *TUI) GetComponent(id int) *Component {
	for _, c := range t.root.Traverse() {
		if c.Id() == id {
			return c
		}
	}
	return nil
}

func (t *TUI) RemoveComponent(id int) error {
	componentToRemove := t.GetComponent(id)
	if componentToRemove == nil {
		return nil
	}
	if !componentToRemove.IsRoot() {
		componentToRemove.Parent().RemoveChild(id)
		componentToRemove.RemoveChildren()
	}
	for _, c := range componentToRemove.Traverse() {
		c.SetParent(nil)
		c.RemoveChildren()
	}
	return nil
}

var floatBoxId int = -1

func toggleFloatBox(t *TUI) {
	var f Component
	if floatBoxId == -1 {
		f = t.NewFloatingBox()
		f.SetStyle(tcell.StyleDefault.Foreground(color.White).Background(color.Green))
		f.SetPadding(2)
		f.Spec().(*Box).SetBorder(BorderStyleRounded)
		f.SetWidth(10)
		f.SetHeight(10)
		f.SetX(10)
		text := t.NewText("Hello World")
		text.SetStyle(tcell.StyleDefault.Foreground(color.White).Background(color.Green))
		f.AddChild(&text)
		t.root.AddChild(&f)
		floatBoxId = f.Id()
		t.UpdateGeometry(&f)
	} else {
		t.RemoveComponent(floatBoxId)
		floatBoxId = -1
		t.UpdateGeometry(nil)
	}
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
	t.root.AddChild(&box1)
	t.root.AddChild(&box2)
	t.root.AddChild(&box3)
	box1.SetMinWidth(30)
	box1.SetMaxWidth(30)
	t.UpdateGeometry(nil)
	t.screen.Show()
	defer t.Quit()
	for {
		t.DrawAll()
		ev := <-t.screen.EventQ()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			t.UpdateGeometry(nil)
			t.screen.Sync()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyCtrlC, tcell.KeyEscape:
				t.Quit()
				return
			case tcell.KeyRight:
				box1.SetMinWidth(box1.MinWidth() + 1)
				box1.SetMaxWidth(box1.MaxWidth() + 1)
				t.UpdateGeometry(nil)
			case tcell.KeyLeft:
				box1.SetMinWidth(box1.MinWidth() - 1)
				box1.SetMaxWidth(box1.MaxWidth() - 1)
				t.UpdateGeometry(nil)
			case tcell.KeyEnter:
				toggleFloatBox(t)
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
