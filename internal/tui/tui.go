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
	"strings"
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

func (t *TUI) CalculateGridGeometry(components []*Component, parentComponent *Component) []int {
	sizes := make([]int, len(components))
	targetSize := parentComponent.Width()
	if parentComponent.Layout() == LayoutVerticalGrid {
		targetSize = parentComponent.Height()
	}
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
	offset := 0
	ids := make([]int, len(components))
	for i, c := range components {
		if parentComponent.Layout() == LayoutHorizontalGrid {
			c.SetGeometry(offset, 0, sizes[i], parentComponent.Height())
		} else {
			c.SetGeometry(0, offset, parentComponent.Width(), sizes[i])
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
				c.SetGeometry(0, 0, c.Parent().Width(), c.Parent().Height())
				processed[c.Id()] = true
			} else if !c.IsRoot() && (c.Parent().Layout() == LayoutHorizontalGrid || c.Parent().Layout() == LayoutVerticalGrid) {
				for _, i := range t.CalculateGridGeometry(c.GetSiblings(), c.Parent()) {
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
		if i == 2 {
			t := t.NewText(0, 0, 0, 0, 1,
				"Everyone is a genius. But if you judge a fish by its ability to climb a tree, it will live its whole life believing that it is stupid.",
				true, TextAlignmentJustify)
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

func (c *Component) Children() []*Component {
	return c.children
}

func (c *Component) SetParent(newParent *Component) {
	c.parent = newParent
}

func (c *Component) Parent() *Component {
	return c.parent
}

func (c *Component) IsRoot() bool {
	return c.Parent() == nil
}

func (c *Component) GetSiblings() []*Component {
	if c.IsRoot() {
		return make([]*Component, 0)
	}
	siblings := c.Parent().Children()
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
	return c.x + c.Parent().AbsX()
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
	return c.y + c.Parent().AbsY()
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

func (c *Component) HasChildren() bool {
	return len(c.children) > 0
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
		for _, c := range (*e).Children() {
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
	s := c.Screen()
	for x := range c.Width() {
		for y := range c.Height() {
			s.Put(c.AbsX()+x, c.AbsY()+y, " ", c.style)
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
	wordWrap  bool
	text      string
}

func (t *TUI) NewText(width int, height int, x int, y int, padding int, text string, wordWrap bool, alignment TextAlignment) Component {
	return t.NewComponent(width, height, x, y, &Text{alignment: alignment, wordWrap: wordWrap, text: text})
}

func (t *Text) SetText(text string) {
	t.text = text
}

func (t *Text) SetAlignment(alignment TextAlignment) {
	t.alignment = alignment
}

func (t *Text) SetWordWrap(wordWrap bool) {
	t.wordWrap = wordWrap
}

func (t *Text) Text() string {
	return t.text
}

func (t *Text) Alignment() TextAlignment {
	return t.alignment
}

func (t *Text) RenderLine(words []string, alignment TextAlignment, allowedWidth int) string {
	if len(words) == 0 {
		return ""
	}
	res := ""
	switch alignment {
	case TextAlignmentLeft:
		res = strings.Join(words, " ")
	case TextAlignmentRight:
		s := strings.Join(words, " ")
		res = strings.Repeat(" ", allowedWidth-len(s)) + s
	case TextAlignmentCenter:
		s := strings.Join(words, " ")
		spacesBefore := (allowedWidth - len(s)) / 2
		spacesAfter := (allowedWidth - len(s)) - spacesBefore
		res = strings.Repeat(" ", spacesBefore) + s + strings.Repeat(" ", spacesAfter)
	case TextAlignmentJustify:
		if len(words) == 1 {
			res = words[0]
			break
		}
		fillLength := allowedWidth - len(strings.Join(words, " "))
		spaces := make([]int, len(words)-1)
		i := 0
		for fillLength > 0 {
			if i == len(spaces)-1 {
				i = 0
			}
			spaces[i] += 1
			fillLength -= 1
			i++
		}
		res = ""
		for i := range words {
			res += words[i]
			if i < len(words)-1 {
				res += strings.Repeat(" ", spaces[i]+1)
			}
		}
	}
	return res
}

func (t *Text) Draw(c *Component) error {
	s := c.Screen()
	lines := make([][]string, 0)
	allowedWidth := c.Width() - 2 // TODO: add dynamic padding
	words := strings.Split(t.text, " ")
	currentLength := 0
	currentLine := make([]string, 0)
	for _, w := range words {
		if currentLength+len(w)+1 > allowedWidth {
			lines = append(lines, currentLine)
			currentLine = make([]string, 0)
			currentLine = append(currentLine, w)
			currentLength = len(w)
		} else {
			currentLine = append(currentLine, w)
			currentLength += len(w) + 1
		}
	}
	if len(currentLine) > 0 {
		lines = append(lines, currentLine)
	}
	allowedHeight := c.Height() - 2 // TODO: add dynamic padding
	for y := range utils.Min(len(lines), allowedHeight) {
		l := t.RenderLine(lines[y], t.alignment, allowedWidth)
		for x := range utils.Min(len(l), allowedWidth) {
			s.Put(c.AbsX()+x+1, c.AbsY()+1+y, l[x:x+1], c.Style())
		}
	}
	return nil
}
