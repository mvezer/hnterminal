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

type TextAlignment int

const (
	TextAlignmentLeft TextAlignment = iota
	TextAlignmentCenter
	TextAlignmentRight
	TextAlignmentJustify
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
	root := tui.NewBox(0, 0, 0, 0, 0)
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

// Traverse through the list of components and updates the geometry for the ones that have changed
// All the components are considered changed in the subtree of a changed component and also in the subtrees of the
// siblings of the changed component
func (t *TUI) UpdateGeometry(rootComponent *Component) {
	if rootComponent == nil {
		rootComponent = t.root
	}
	if !rootComponent.IsRoot() && (rootComponent.Layout() == LayoutVerticalGrid || rootComponent.Layout() == LayoutHorizontalGrid) {
		rootComponent = rootComponent.GetParent()
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
			parentLayout := LayoutFill   // default layout
			parentWidth := screenWidth   // default width (screen width)
			parentHeight := screenHeight // default height (screen height)
			if !c.IsRoot() {             // if not root
				parentLayout = c.GetParent().Layout()
				parentWidth, parentHeight = c.GetParent().Size()
			}
			switch parentLayout {
			case LayoutFill:
				c.SetGeometry(0, 0, parentWidth, parentHeight)
				processed[c.Id()] = true
			case LayoutVerticalGrid:
				siblings := c.GetSiblings()
				h := 0
				y := 0
				for i, s := range siblings {
					if i == len(siblings)-1 { // the last sibling
						h = parentHeight - (len(siblings)-1)*h
					} else {
						h = parentHeight / len(siblings)
					}
					s.SetGeometry(0, y, parentWidth, h)
					processed[s.Id()] = true
					y += h
				}
			case LayoutHorizontalGrid:
				siblings := c.GetSiblings()
				parentWidth, parentHeight = c.GetParent().Size()
				w := 0
				x := 0
				for i, s := range siblings {
					if i == len(siblings)-1 { // the last sibling
						w = parentWidth - (len(siblings)-1)*w
					} else {
						w = parentWidth / len(siblings)
					}
					s.SetGeometry(x, 0, w, parentHeight)
					processed[s.Id()] = true
					x += w
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

func (t *TUI) NewComponent(width int, height int, x int, y int, spec ComponentSpec) Component {
	return Component{
		id:        t.NextId(),
		tui:       t,
		width:     width,
		height:    height,
		minWidth:  -1,
		maxWidth:  -1,
		minHeight: -1,
		maxHeight: -1,
		x:         x,
		y:         y,
		children:  make([]*Component, 0),
		parent:    nil,
		style:     t.defaultStyle,
		layout:    LayoutFill,
		floating:  false,
		zIndex:    -1,
		spec:      spec,
	}
}

func (t *TUI) Run() {
	box1 := t.NewBox(0, 0, 0, 0, 0)
	box1.SetStyle(tcell.StyleDefault.Foreground(color.White).Background(color.Green))
	box2 := t.NewBox(0, 0, 0, 0, 0)
	box2.SetLayout(LayoutVerticalGrid)
	box2.SetStyle(tcell.StyleDefault.Foreground(color.Red).Background(color.LightCyan))
	box3 := t.NewBox(0, 0, 0, 0, 0)
	box3.SetLayout(LayoutVerticalGrid)
	box3.SetStyle(tcell.StyleDefault.Foreground(color.White).Background(color.Red))
	for i := range 5 {
		b := t.NewBox(0, 0, 0, 0, 0)
		if i%2 == 0 {
			b.SetStyle(tcell.StyleDefault.Foreground(color.White).Background(color.Black))
		} else {
			b.SetStyle(tcell.StyleDefault.Foreground(color.Black).Background(color.White))
		}
		box2.AddChild(&b)
	}
	t.root.SetLayout(LayoutHorizontalGrid)
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

// ----------------- Component -----------------
type ComponentSpec interface {
	Draw(*Component) error
}

type Component struct {
	id        int
	tui       *TUI
	width     int
	height    int
	minWidth  int
	minHeight int
	maxWidth  int
	maxHeight int
	x         int
	y         int
	children  []*Component
	parent    *Component
	style     tcell.Style
	layout    Layout
	floating  bool
	zIndex    int
	spec      ComponentSpec
}

func (c *Component) AddChild(newChild *Component) error {
	c.tui.mutex.Lock()
	newChild.SetZIndex(c.zIndex + 1)
	newChild.SetParent(c)
	c.children = append(c.children, newChild)
	c.tui.mutex.Unlock()
	return nil
}

func (c *Component) RemoveChild(id int) error {
	for i, child := range c.children {
		if child.Id() == id {
			c.children = append(c.children[:i], c.children[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("element (%d) not found", id)
}

func (c *Component) GetChildren() []*Component {
	return c.children
}

func (c *Component) SetParent(newParent *Component) {
	c.parent = newParent
}

func (c *Component) GetParent() *Component {
	return c.parent
}

func (c *Component) IsRoot() bool {
	return c.GetParent() == nil
}

func (c *Component) GetSiblings() []*Component {
	if c.IsRoot() {
		return make([]*Component, 0)
	}
	siblings := c.GetParent().GetChildren()
	siblingsCopy := make([]*Component, len(siblings))
	copy(siblingsCopy, siblings)
	slices.SortFunc(siblingsCopy, func(a, b *Component) int {
		return a.Id() - b.Id()
	})
	return siblingsCopy
}

func (c *Component) Style() tcell.Style {
	return c.style
}

func (c *Component) SetStyle(style tcell.Style) {
	c.tui.mutex.Lock()
	c.style = style
	c.tui.mutex.Unlock()
}

func (c *Component) Width() int {
	return c.width
}

func (c *Component) MinWidth() int {
	return c.minWidth
}

func (c *Component) MaxWidth() int {
	return c.maxWidth
}

func (c *Component) MinHeight() int {
	return c.minHeight
}

func (c *Component) MaxHeight() int {
	return c.maxHeight
}

func (c *Component) SetMinWidth(minWidth int) {
	c.minWidth = minWidth
}

func (c *Component) SetMaxWidth(maxWidth int) {
	c.maxWidth = maxWidth
}

func (c *Component) SetMinHeight(minHeight int) {
	c.minHeight = minHeight
}

func (c *Component) SetMaxHeight(maxHeight int) {
	c.maxHeight = maxHeight
}

func (c *Component) SetWidth(width int) bool {
	if c.width == width {
		return false
	}
	c.tui.mutex.Lock()
	c.width = width
	c.tui.mutex.Unlock()
	return true
}

func (c *Component) Height() int {
	return c.height
}

func (c *Component) SetHeight(height int) bool {
	if c.height == height {
		return false
	}
	c.height = height
	return true
}

func (c *Component) X() int {
	return c.x
}
func (c *Component) AbsX() int {
	if c.IsRoot() {
		return c.x
	}
	return c.x + c.GetParent().AbsX()
}

func (c *Component) SetX(x int) bool {
	if c.x == x {
		return false
	}
	c.x = x
	return true
}

func (c *Component) Y() int {
	return c.y
}

func (c *Component) AbsY() int {
	if c.IsRoot() {
		return c.y
	}
	return c.y + c.GetParent().AbsY()
}

func (c *Component) SetY(y int) bool {
	if c.y == y {
		return false
	}
	c.y = y
	return true
}

func (c *Component) Screen() tcell.Screen {
	return c.tui.screen
}

func (c *Component) Id() int {
	return c.id
}

func (c *Component) Root() *Component {
	return c.tui.root
}

func (c *Component) Spec() ComponentSpec {
	return c.spec
}

func (c *Component) Draw() error {
	return c.spec.Draw(c)
}

func (c *Component) Layout() Layout {
	return c.layout
}

func (c *Component) SetLayout(layout Layout) {
	c.layout = layout
}

func (c *Component) Floating() bool {
	return c.floating
}

func (c *Component) SetZIndex(zIndex int) {
	c.zIndex = zIndex
}

func (c *Component) ZIndex() int {
	return c.zIndex
}

func (c *Component) SetGeometry(x int, y int, width int, height int) bool {
	geometryChanged := c.SetX(x)
	geometryChanged = c.SetY(y) || geometryChanged
	geometryChanged = c.SetWidth(width) || geometryChanged
	geometryChanged = c.SetHeight(height) || geometryChanged
	if geometryChanged {
		c.tui.AddToDrawMap(c)
	}
	return geometryChanged
}

func (c *Component) Traverse() []*Component {
	result := make([]*Component, 0)
	result = append(result, c)
	i := 0
	for i < len(result) {
		e := result[i]
		i++
		for _, c := range (*e).GetChildren() {
			result = append(result, c)
		}
	}
	return result
}

func (c *Component) Size() (int, int) {
	return c.width, c.height
}

func (c *Component) SetSize(width int, height int) {
	c.width = width
	c.height = height
}

// ----------------- Box -----------------
type Box struct {
	Padding int
}

func (b Box) Draw(c *Component) error {
	c.tui.mutex.Lock()
	log.Printf("Drawing box %d, x: %d, y: %d, w: %d, h: %d", c.Id(), c.AbsX(), c.AbsY(), c.Width(), c.Height())
	s := c.Screen()
	for x := range c.Width() {
		for y := range c.Height() {
			s.Put(c.AbsX()+x, c.AbsY()+y, "O", c.style)
		}
	}
	c.tui.mutex.Unlock()
	return nil
}
func (t *TUI) NewBox(width int, height int, x int, y int, padding int) Component {
	return t.NewComponent(width, height, x, y, Box{Padding: padding})
}

// ----------------- Text -----------------
type Text struct {
	alignment TextAlignment
	text      string
}

func (t Text) Draw(c *Component) error {
	s := c.Screen()
	s.Put(c.AbsX(), c.AbsY(), t.text, c.Style())
	return nil
}
