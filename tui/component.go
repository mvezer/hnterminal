package tui

import (
	"fmt"
	"hnterminal/utils"
	"iter"
	"slices"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

var DEFAULT_STYLE = tcell.StyleDefault.Background(color.Reset).Foreground(color.White)

type Component interface {
	Id() int
	Draw() error
	BeforeUpdate() error
	AfterUpdate() error
	X() int
	Y() int
	AbsoluteX() int
	AbsoluteY() int
	SetX(int)
	SetY(int)
	Width() int
	Height() int
	FixedWidth() int
	FixedHeight() int
	SetFixedWidth(int)
	SetFixedHeight(int)
	SetWidth(int)
	SetHeight(int)
	SetGeometry(int, int, int, int)
	WidthPercent() int
	HeightPercent() int
	SetWidthPercent(int)
	SetHeightPercent(int)
	Padding() Padding
	SetPadding(Padding)
	Parent() Component
	IsRoot() bool
	SetParent(Component)
	Children() []Component
	AddChild(Component)
	Traverse(filterFunc func(Component) bool) iter.Seq[Component]
	TraverseDFS() iter.Seq2[int, Component]
	Layout() Layout
	SetLayout(Layout)
	Floating() bool
	SetFloating(bool)
	Dirty() bool
	SetDirty(bool)
	Style() tcell.Style
	SetStyle(tcell.Style)
	String() string
}

type ComponentChanged interface {
	tcell.Event
	Component() *BaseComponent
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
	Component
	id            int
	tui           *TUI
	x             int
	y             int
	width         int
	height        int
	widthPercent  int
	heightPercent int
	style         tcell.Style
	children      []Component
	parent        Component
	layout        Layout
	floating      bool
	fixedWidth    int
	fixedHeight   int
	padding       Padding
	dirty         bool
}

func (bc *BaseComponent) Id() int {
	return bc.id
}

func (bc *BaseComponent) ResetGeometry() {
	bc.width = -1
	bc.height = -1
	bc.dirty = true
}

func (bc *BaseComponent) SetPadding(p Padding) {
	bc.padding = p
}

func (bc *BaseComponent) Padding() Padding {
	return bc.padding
}

func (bc *BaseComponent) SetLayout(l Layout) {
	bc.layout = l
	bc.dirty = true
}

func (bc *BaseComponent) AbsoluteX() int {
	if bc.parent == nil {
		return bc.x
	}
	return bc.x + bc.parent.AbsoluteX()
}

func (bc *BaseComponent) AbsoluteY() int {
	if bc.parent == nil {
		return bc.y
	}
	return bc.y + bc.parent.AbsoluteY()
}

func (bc *BaseComponent) Parent() Component {
	return bc.parent
}

func (bc *BaseComponent) SetParent(parent Component) {
	bc.parent = parent
}

func (bc *BaseComponent) Children() []Component {
	return bc.children
}

func (bc *BaseComponent) FilteredChildren(filter func(Component) bool) []Component {
	return slices.Collect(func(yield func(Component) bool) {
		for _, child := range bc.children {
			if filter(child) {
				if !yield(child) {
					return
				}
			}
		}
	})
}

func (bc *BaseComponent) AddChild(child Component) {
	// TODO: check if child already has a parent
	// TODO: check if the child is already added to this component
	bc.children = append(bc.children, child)
	child.SetParent(bc)
}

func (bc *BaseComponent) HasChildren() bool {
	return len(bc.Children()) > 0
}

func (bc *BaseComponent) RemoveChildById(id int) {
	for i, child := range bc.children {
		if child.Id() == id {
			bc.children = append(bc.children[:i], bc.children[i+1:]...)
			break
		}
	}
}

func (bc *BaseComponent) Style() tcell.Style {
	return bc.style
}

func (bc *BaseComponent) SetStyle(style tcell.Style) {
	bc.style = style
}

func (bc *BaseComponent) Width() int {
	return bc.width
}

func (bc *BaseComponent) SetWidth(width int) {
	if bc.width != width {
		bc.width = width
		bc.dirty = true
	}
}

func (bc *BaseComponent) Height() int {
	return bc.height
}

func (bc *BaseComponent) SetHeight(height int) {
	if bc.height != height {
		bc.height = height
		bc.dirty = true
	}
}

func (bc *BaseComponent) X() int {
	return bc.x
}

func (bc *BaseComponent) SetX(x int) {
	if bc.x != x {
		bc.x = x
		bc.dirty = true
	}
}

func (bc *BaseComponent) Y() int {
	return bc.y
}

func (bc *BaseComponent) SetY(y int) {
	if bc.y != y {
		bc.y = y
		bc.dirty = true
	}
}

func (bc *BaseComponent) Floating() bool {
	return bc.floating
}

func (bc *BaseComponent) HeightPercent() int {
	return bc.heightPercent
}

func (bc *BaseComponent) WidthPercent() int {
	return bc.widthPercent
}

func (bc *BaseComponent) SetWidthPercent(widthPercent int) {
	if bc.widthPercent != widthPercent {
		bc.widthPercent = widthPercent
		bc.dirty = true
	}
}

func (bc *BaseComponent) SetHeightPercent(heightPercent int) {
	bc.heightPercent = heightPercent
}

func (bc *BaseComponent) FixedWidth() int {
	return bc.fixedWidth
}

func (bc *BaseComponent) FixedHeight() int {
	return bc.fixedHeight
}

func (bc *BaseComponent) SetFixedWidth(width int) {
	bc.fixedWidth = width
}

func (bc *BaseComponent) SetFixedHeight(height int) {
	bc.fixedHeight = height
}

func (bc *BaseComponent) SetGeometry(x int, y int, w int, h int) {
	bc.SetX(x)
	bc.SetY(y)
	bc.SetWidth(w)
	bc.SetHeight(h)
}

func (bc *BaseComponent) Dirty() bool {
	return bc.dirty
}

func (bc *BaseComponent) SetDirty(dirty bool) {
	bc.dirty = dirty
}

func (bc *BaseComponent) Draw() error {
	bc.dirty = false
	return nil
}

func (bc *BaseComponent) String() string {
	layoutString := ""
	switch bc.layout {
	case HorizontalGrid:
		layoutString = "HorizontalGrid"
	case VerticalGrid:
		layoutString = "VerticalGrid"
	case FixedWidth:
		layoutString = "FixedWidth"
	case FixedHeight:
		layoutString = "FixedHeight"
	}
	return fmt.Sprintf("Component #%d [%s]\nx: %d, y: %d\nlayout: %s,\nwidth: %d, height: %d\nfixedWidth: %d, fixedHeight: %d\nfloating: %t\npadding: %+v\n", bc.id, bc.String(), bc.x, bc.y, layoutString, bc.width, bc.height, bc.fixedWidth, bc.fixedHeight, bc.floating, bc.padding)
}

func (bc *BaseComponent) IsRoot() bool {
	return bc.parent == nil
}

func (bc *BaseComponent) Find(findFunc func(Component) bool) Component {
	for c := range bc.Traverse(nil) {
		if findFunc(c) {
			return c
		}
	}
	return nil
}

func (bc *BaseComponent) Traverse(filterFunc func(Component) bool) iter.Seq[Component] {
	return func(yield func(Component) bool) {
		fifo := utils.FIFO[Component]{}
		fifo.Enqueue(bc)

		for !fifo.IsEmpty() {
			e := fifo.Dequeue()
			if filterFunc != nil && !filterFunc(e) {
				continue
			}
			if !yield(e) {
				return
			}
			for _, c := range e.Children() {
				fifo.Enqueue(c)
			}
		}
	}
}

func (bc *BaseComponent) TraverseDFS() iter.Seq2[int, Component] {
	return func(yield func(int, Component) bool) {
		type ComponentWithDepth struct {
			depth int
			c     Component
		}
		stack := utils.Stack[ComponentWithDepth]{}
		stack.Push(ComponentWithDepth{0, bc})

		for !stack.IsEmpty() {
			cwd := stack.Pop()
			if !yield(cwd.depth, cwd.c) {
				return
			}
			for _, c := range cwd.c.Children() {
				stack.Push(ComponentWithDepth{cwd.depth + 1, c})
			}
		}
	}
}

// Traverses all the componetest in the subtree of the component
// filters out all floating components (except the root)
// func (bc *BaseComponent) TraverseSubtree() iter.Seq[*BaseComponent] {
// 	return func(yield func(*BaseComponent) bool) {
// 		fifo := utils.FIFO[BaseComponent]{}
// 		fifo.Enqueue(bc)
//
// 		for !fifo.IsEmpty() {
// 			e := fifo.Dequeue()
// 			if !yield(e) {
// 				return
// 			}
// 			for _, c := range e.Children() {
// 				if !c.Floating() {
// 					fifo.Enqueue(c)
// 				}
// 			}
// 		}
// 	}
// }

// func (bc *BaseComponent) TraverseFloating() iter.Seq[*BaseComponent] {
// 	return func(yield func(*BaseComponent) bool) {
// 		fifo := utils.FIFO[BaseComponent]{}
// 		fifo.Enqueue(bc)
//
// 		for !fifo.IsEmpty() {
// 			e := fifo.Dequeue()
// 			if e.floating {
// 				if !yield(e) {
// 					return
// 				}
// 			}
// 			for _, c := range e.Children() {
// 				fifo.Enqueue(c)
// 			}
// 		}
// 	}
// }

// func (bc *BaseComponent) DebugTree() string {
// 	builder := strings.Builder{}
// 	for depth, c := range bc.TraverseDFS() {
// 		builder.WriteString(strings.Repeat("  ", depth*2))
// 		fmt.Fprintf(&builder, "[%d]. (%s) x: %d, y: %d, width: %d, height: %d, fixedWidth: %d, fixedHeight: %d\n", c.Id(), c.String(), c., c.y, c.width, c.height, c.fixedWidth, c.fixedHeight)
// 	}
// 	return builder.String()
// }
