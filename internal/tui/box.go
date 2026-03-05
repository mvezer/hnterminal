package tui

import (
	"fmt"
)

type BorderStyle int

const (
	BorderStyleNone BorderStyle = iota
	BorderStyleSingle
	BorderStyleDouble
	BorderStyleThick
	BorderStyleRounded
)

const DEFAULT_BORDER_STYLE = BorderStyleSingle

type Border struct {
	Left   bool
	Top    bool
	Right  bool
	Bottom bool
}

type BorderElements struct {
	horizontal  rune
	vertical    rune
	topLeft     rune
	topRight    rune
	bottomLeft  rune
	bottomRight rune
}

var borders = map[BorderStyle]BorderElements{
	BorderStyleSingle: {
		horizontal:  '─',
		vertical:    '│',
		topLeft:     '┌',
		topRight:    '┐',
		bottomLeft:  '└',
		bottomRight: '┘',
	},
	BorderStyleRounded: {
		horizontal:  '─',
		vertical:    '│',
		topLeft:     '╭',
		topRight:    '╮',
		bottomLeft:  '╰',
		bottomRight: '╯',
	},
	BorderStyleThick: {
		horizontal:  '━',
		vertical:    '┃',
		topLeft:     '┏',
		topRight:    '┓',
		bottomLeft:  '┗',
		bottomRight: '┛',
	},
	BorderStyleDouble: {
		horizontal:  '═',
		vertical:    '║',
		topLeft:     '╔',
		topRight:    '╗',
		bottomLeft:  '╚',
		bottomRight: '╝',
	},
}

type Box struct {
	borderStyle BorderStyle
	border      Border
}

type TextAlignment int

const (
	TextAlignLeft TextAlignment = iota
	TextAlignCenter
	TextAlignRight
	TextAlignJustify
)

func NewBox(layout Layout) BaseComponent {
	b := Box{}
	return NewComponent(&b, layout)
}

func NewFloatingBox(layout Layout) BaseComponent {
	b := Box{}
	c := NewComponent(&b, layout)
	c.floating = true
	return c
}

func (b *Box) OnUpdate(c *BaseComponent) error {
	return nil
}

func (b Box) Draw(c *BaseComponent, tui *TUI) error {
	if c.width <= 0 || c.height <= 0 {
		return nil
	}
	for y := 0; y < c.height; y++ {
		for x := 0; x < c.width; x++ {
			chr := ' '
			if b.borderStyle != BorderStyleNone {
				top := y == 0
				bottom := y == c.height-1
				left := x == 0
				right := x == c.width-1
				if top && left {
					if b.border.Top && b.border.Left {
						chr = borders[b.borderStyle].topLeft
					} else if b.border.Top {
						chr = borders[b.borderStyle].horizontal
					} else if b.border.Left {
						chr = borders[b.borderStyle].vertical
					}
				} else if top && right {
					if b.border.Top && b.border.Right {
						chr = borders[b.borderStyle].topRight
					} else if b.border.Top {
						chr = borders[b.borderStyle].horizontal
					} else if b.border.Right {
						chr = borders[b.borderStyle].vertical
					}
				} else if bottom && left {
					if b.border.Bottom && b.border.Left {
						chr = borders[b.borderStyle].bottomLeft
					} else if b.border.Bottom {
						chr = borders[b.borderStyle].horizontal
					} else if b.border.Left {
						chr = borders[b.borderStyle].vertical
					}
				} else if bottom && right {
					if b.border.Bottom && b.border.Right {
						chr = borders[b.borderStyle].bottomRight
					} else if b.border.Bottom {
						chr = borders[b.borderStyle].horizontal
					} else if b.border.Right {
						chr = borders[b.borderStyle].vertical
					}
				} else if top {
					if b.border.Top {
						chr = borders[b.borderStyle].horizontal
					}
				} else if bottom {
					if b.border.Bottom {
						chr = borders[b.borderStyle].horizontal
					}
				} else if left {
					if b.border.Left {
						chr = borders[b.borderStyle].vertical
					}
				} else if right {
					if b.border.Right {
						chr = borders[b.borderStyle].vertical
					}
				}
			}

			tui.screen.SetContent(c.AbsoluteX()+x, c.AbsoluteY()+y, chr, nil, c.style)
		}
	}
	return nil
}

func (b *Box) SetBorderStyle(borderStyle BorderStyle) {
	b.borderStyle = borderStyle
}

func (b *Box) SetBorder(border Border) {
	b.border = border
}

func (b Box) String() string {
	borderStyle := "none"
	switch b.borderStyle {
	case BorderStyleNone:
		borderStyle = "none"
	case BorderStyleSingle:
		borderStyle = "single"
	case BorderStyleDouble:
		borderStyle = "double"
	case BorderStyleThick:
		borderStyle = "thick"
	case BorderStyleRounded:
		borderStyle = "rounded"
	}
	return fmt.Sprintf("Box (border: %s)", borderStyle)
}
