package tui

type BorderStyle int

const (
	BorderStyleNone BorderStyle = iota
	BorderStyleSingle
	BorderStyleThick
	BorderStyleRounded
	BorderStyleDouble
)

type BorderElements struct {
	horizontal  string
	vertical    string
	topLeft     string
	topRight    string
	bottomLeft  string
	bottomRight string
}

var borders = map[BorderStyle]BorderElements{
	BorderStyleSingle: BorderElements{
		horizontal:  "─",
		vertical:    "│",
		topLeft:     "┌",
		topRight:    "┐",
		bottomLeft:  "└",
		bottomRight: "┘",
	},
	BorderStyleRounded: BorderElements{
		horizontal:  "─",
		vertical:    "│",
		topLeft:     "╭",
		topRight:    "╮",
		bottomLeft:  "╰",
		bottomRight: "╯",
	},
	BorderStyleThick: BorderElements{
		horizontal:  "━",
		vertical:    "┃",
		topLeft:     "┏",
		topRight:    "┓",
		bottomLeft:  "┗",
		bottomRight: "┛",
	},
	BorderStyleDouble: BorderElements{
		horizontal:  "═",
		vertical:    "║",
		topLeft:     "╔",
		topRight:    "╗",
		bottomLeft:  "╚",
		bottomRight: "╝",
	},
}

type Box struct {
	padding int
	border  BorderStyle
}

func (b Box) Draw(c *Component) error {
	s := c.Screen()
	for x := range c.Width() {
		for y := range c.Height() {
			s.Put(c.AbsX()+x, c.AbsY()+y, " ", c.style)
			if b.border != BorderStyleNone {
				if x == 0 && y == 0 {
					s.Put(c.AbsX()+x, c.AbsY()+y, borders[b.border].topLeft, c.style)
				} else if x == 0 && y == c.Height()-1 {
					s.Put(c.AbsX()+x, c.AbsY()+y, borders[b.border].bottomLeft, c.style)
				} else if x == c.Width()-1 && y == 0 {
					s.Put(c.AbsX()+x, c.AbsY()+y, borders[b.border].topRight, c.style)
				} else if x == c.Width()-1 && y == c.Height()-1 {
					s.Put(c.AbsX()+x, c.AbsY()+y, borders[b.border].bottomRight, c.style)
				} else if y == 0 || y == c.Height()-1 {
					s.Put(c.AbsX()+x, c.AbsY()+y, borders[b.border].horizontal, c.style)
				} else if x == 0 || x == c.Width()-1 {
					s.Put(c.AbsX()+x, c.AbsY()+y, borders[b.border].vertical, c.style)
				}
			}
		}
	}
	return nil
}

func (b *Box) Border() BorderStyle {
	return b.border
}
func (b *Box) SetBorder(border BorderStyle) {
	b.border = border
}

func (t *TUI) NewBox() Component {
	return t.NewComponent(&Box{padding: 0, border: BorderStyleNone})
}
