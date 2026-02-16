package tui

import (
	"hnterminal/internal/utils"
	"strings"
)

type TextAlignment int

const (
	TextAlignmentLeft TextAlignment = iota
	TextAlignmentCenter
	TextAlignmentRight
	TextAlignmentJustify
)

type Text struct {
	alignment TextAlignment
	wordWrap  bool
	text      string
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

func (t *TUI) NewText(text string) Component {
	return t.NewComponent(&Text{alignment: TextAlignmentLeft, wordWrap: true, text: text}, false)
}
