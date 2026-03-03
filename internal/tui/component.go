package tui

import (
	"fmt"
	"hnterminal/internal/utils"
	"iter"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

var DEFAULT_STYLE = tcell.StyleDefault.Background(color.Reset).Foreground(color.White)

type Component interface {
	Draw(*BaseComponent, *TUI) error
	OnUpdate(*BaseComponent) error
	String() string
}

type Padding struct {
	Left   int
	Top    int
	Right  int
	Bottom int
}

var lastComponentId int = -1

func getNextComponentId() int {
	lastComponentId++
	return lastComponentId
}

type BaseComponent struct {
	id            int
	x             int
	y             int
	width         int
	height        int
	widthPercent  int
	heightPercent int
	style         tcell.Style
	children      []*BaseComponent
	parent        *BaseComponent
	layout        Layout
	kind          Component
	floating      bool
	fixedWidth    int
	fixedHeight   int
	padding       Padding
	dirty         bool
}

func (c *BaseComponent) Id() int {
	return c.id
}

func (c *BaseComponent) ResetGeometry() {
	c.width = -1
	c.height = -1
	c.dirty = true
}

func (c *BaseComponent) SetPadding(p Padding) {
	c.padding = p
}

func (c *BaseComponent) Padding() Padding {
	return c.padding
}

func (c *BaseComponent) SetLayout(l Layout) {
	c.layout = l
}

func (c *BaseComponent) AbsoluteX() int {
	if c.parent == nil {
		return c.x
	}
	return c.x + c.parent.AbsoluteX()
}

func (c *BaseComponent) AbsoluteY() int {
	if c.parent == nil {
		return c.y
	}
	return c.y + c.parent.AbsoluteY()
}

func (c *BaseComponent) Parent() *BaseComponent {
	return c.parent
}

func (c *BaseComponent) SetParent(parent *BaseComponent) {
	c.parent = parent
}

func (c *BaseComponent) Children() []*BaseComponent {
	return c.children
}

func (c *BaseComponent) AddChild(child *BaseComponent) {
	// TODO: check if child already has a parent
	// TODO: check if the child is already added to this component
	c.children = append(c.children, child)
	child.SetParent(c)
}

func (c *BaseComponent) HasChildren() bool {
	return len(c.Children()) > 0
}

func (c *BaseComponent) RemoveChildById(id int) {
	for i, child := range c.children {
		if child.Id() == id {
			c.children = append(c.children[:i], c.children[i+1:]...)
			break
		}
	}
}

func (c *BaseComponent) Style() tcell.Style {
	return c.style
}

func (c *BaseComponent) SetStyle(style tcell.Style) {
	c.style = style
}

func (c *BaseComponent) Width() int {
	return c.width
}

func (c *BaseComponent) SetWidth(width int) {
	if c.width != width {
		c.width = width
		c.dirty = true
		c.kind.OnUpdate(c)
	}
}

func (c *BaseComponent) Height() int {
	return c.height
}

func (c *BaseComponent) SetHeight(height int) {
	if c.height != height {
		c.height = height
		c.dirty = true
		c.kind.OnUpdate(c)
	}
}

func (c *BaseComponent) X() int {
	return c.x
}

func (c *BaseComponent) SetX(x int) {
	if c.x != x {
		c.x = x
		c.dirty = true
		c.kind.OnUpdate(c)
	}
}

func (c *BaseComponent) Y() int {
	return c.y
}

func (c *BaseComponent) SetY(y int) {
	if c.y != y {
		c.y = y
		c.dirty = true
		c.kind.OnUpdate(c)
	}
}

func (c *BaseComponent) SetGeometry(x, y, width, height int) bool {
	c.SetX(x)
	c.SetY(y)
	c.SetWidth(width)
	c.SetHeight(height)
	return c.dirty
}

func (c *BaseComponent) Size() (int, int) {
	return c.width, c.height
}

func (c *BaseComponent) SetSize(width int, height int) {
	c.width = width
	c.height = height
}

func (c *BaseComponent) Floating() bool {
	return c.floating
}

func (c *BaseComponent) HeightPercent() int {
	return c.heightPercent
}

func (c *BaseComponent) WidthPercent() int {
	return c.widthPercent
}

func (c *BaseComponent) SetWidthPercent(widthPercent int) {
	if c.widthPercent != widthPercent {
		c.widthPercent = widthPercent
		c.dirty = true
	}
}

func (c *BaseComponent) SetHeightPercent(heightPercent int) {
	c.heightPercent = heightPercent
}

func (c *BaseComponent) FixedWidth() int {
	return c.fixedWidth
}

func (c *BaseComponent) FixedHeight() int {
	return c.fixedHeight
}

func (c *BaseComponent) SetFixedWidth(width int) {
	c.fixedWidth = width
}

func (c *BaseComponent) SetFixedHeight(height int) {
	c.fixedHeight = height
}

func (c *BaseComponent) Dirty() bool {
	return c.dirty
}

func (c *BaseComponent) SetDirty(dirty bool) {
	c.dirty = dirty
}

func (c *BaseComponent) Draw(t *TUI) {
	c.kind.Draw(c, t)
	c.dirty = false
}

func (c *BaseComponent) String() string {
	layoutString := ""
	switch c.layout {
	case HorizontalGrid:
		layoutString = "HorizontalGrid"
	case VerticalGrid:
		layoutString = "VerticalGrid"
	case FixedWidth:
		layoutString = "FixedWidth"
	case FixedHeight:
		layoutString = "FixedHeight"
	}
	return fmt.Sprintf("Component #%d [%s]\nx: %d, y: %d\nlayout: %s\n,width: %d, height: %d\nfixedWidth: %d, fixedHeight: %d\nfloating: %t\n", c.id, c.kind.String(), c.x, c.y, layoutString, c.width, c.height, c.fixedWidth, c.fixedHeight, c.floating)
}

func NewComponent(kind Component, layout Layout) BaseComponent {
	return BaseComponent{
		id:          getNextComponentId(),
		style:       DEFAULT_STYLE,
		kind:        kind,
		floating:    false,
		layout:      layout,
		fixedWidth:  -1,
		fixedHeight: -1,
		width:       -1,
		height:      -1,
		padding:     Padding{0, 0, 0, 0},
	}
}

func (c *BaseComponent) IsRoot() bool {
	return c.parent == nil
}

func (c *BaseComponent) Find(findFunc func(*BaseComponent) bool) *BaseComponent {
	for c := range c.Traverse() {
		if findFunc(c) {
			return c
		}
	}
	return nil
}

func (c *BaseComponent) Traverse() iter.Seq[*BaseComponent] {
	return func(yield func(*BaseComponent) bool) {
		fifo := utils.FIFO[BaseComponent]{}
		fifo.Enqueue(c)

		for !fifo.IsEmpty() {
			e := fifo.Dequeue()
			if !yield(e) {
				return
			}
			for _, c := range e.Children() {
				fifo.Enqueue(c)
			}
		}
	}
}

func (c *BaseComponent) TraverseNonFloating() iter.Seq[*BaseComponent] {
	return func(yield func(*BaseComponent) bool) {
		fifo := utils.FIFO[BaseComponent]{}
		fifo.Enqueue(c)

		for !fifo.IsEmpty() {
			e := fifo.Dequeue()
			if !yield(e) {
				return
			}
			for _, c := range e.Children() {
				if !c.Floating() {
					fifo.Enqueue(c)
				}
			}
		}
	}
}

func (c *BaseComponent) TraverseFloating() iter.Seq[*BaseComponent] {
	return func(yield func(*BaseComponent) bool) {
		fifo := utils.FIFO[BaseComponent]{}
		fifo.Enqueue(c)

		for !fifo.IsEmpty() {
			e := fifo.Dequeue()
			if e.floating {
				if !yield(e) {
					return
				}
			}
			for _, c := range e.Children() {
				fifo.Enqueue(c)
			}
		}
	}
}
