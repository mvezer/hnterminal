package tui

import (
	"fmt"
	"github.com/gdamore/tcell/v3"
	"slices"
)

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
	padding   int
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

func (c *Component) Padding() int {
	return c.padding
}

func (c *Component) SetPadding(padding int) {
	c.padding = padding
}

func (t *TUI) NewComponent(spec ComponentSpec) Component {
	return Component{
		id:        t.NextId(),
		tui:       t,
		width:     0,
		height:    0,
		minWidth:  -1,
		maxWidth:  -1,
		minHeight: -1,
		maxHeight: -1,
		x:         0,
		y:         0,
		children:  make([]*Component, 0),
		parent:    nil,
		style:     t.defaultStyle,
		layout:    LayoutFill,
		floating:  false,
		zIndex:    -1,
		spec:      spec,
		padding:   0,
	}
}
