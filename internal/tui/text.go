package tui

import (
	"fmt"
	"strings"
)

type Text struct {
	text        string
	alignment   TextAlignment
	wrappedText [][]string // text separated to lines and words
}

func (t *Text) SetAlignment(a TextAlignment) {
	t.alignment = a
}

func (t *Text) calculateWordWrap(width int) [][]string {
	words := make([][]string, 0)
	line := make([]string, 0)
	lineLength := 0
	for w := range strings.SplitSeq(t.text, " ") {
		if w == "" || w == " " {
			continue
		}
		lineBreakWordParts := strings.Split(w, "\n")
		for i, w := range lineBreakWordParts {
			if len(w) > width {
				bigWordIdx := 0
				for bigWordIdx < len(w) {
					wordFragment := w[bigWordIdx:min(bigWordIdx+width-lineLength, len(w))]
					bigWordIdx += len(wordFragment)
					if len(line) > 0 {
						line = append(line, wordFragment)
					} else {
						line = make([]string, 1)
						line[0] = wordFragment
					}
					lineLength += len(wordFragment)
					if lineLength == width {
						words = append(words, line)
						line = make([]string, 0)
						lineLength = 0
					}
				}
				continue
			}
			if lineLength+len(w) > width {
				// we need a new line
				words = append(words, line)
				line = make([]string, 1)
				line[0] = w
				lineLength = len(w)
			} else {
				lineLength += len(w) + 1
				line = append(line, w)
			}
			if len(lineBreakWordParts) > 1 && i < len(lineBreakWordParts)-1 {
				words = append(words, line)
				line = make([]string, 0)
				lineLength = 0
			}
		}
	}
	if len(line) > 0 {
		words = append(words, line)
	}

	return words
}

func (t *Text) RenderLine(line []string, width int) string {
	renderedLine := ""
	switch t.alignment {
	case TextAlignLeft:
		renderedLine = strings.Join(line, " ")
	case TextAlignCenter:
		renderedLine = strings.Join(line, " ")
		paddingLeft := strings.Repeat(" ", (width-len(renderedLine))/2)
		paddingRight := strings.Repeat(" ", width-len(renderedLine)-len(paddingLeft))
		renderedLine = paddingLeft + renderedLine + paddingRight
	case TextAlignRight:
		renderedLine = strings.Join(line, " ")
		renderedLine = strings.Repeat(" ", width-len(renderedLine)) + renderedLine
	case TextAlignJustify:
		spaces := make([]string, len(line)-1)
		for i := range spaces {
			spaces[i] = " "
		}
		fill := width
		for _, w := range line {
			fill -= len(w)
		}
		fill -= len(spaces)
		i := 0
		for fill > 0 {
			if i >= len(spaces) {
				i = 0
			}
			spaces[i] += " "
			i++
			fill -= 1
		}
		for i, w := range line {
			renderedLine += w
			if i < len(line)-1 {
				renderedLine += spaces[i]
			}
		}
	}
	return renderedLine
}

func (t Text) Draw(c *BaseComponent, tui *TUI) error {
	if t.wrappedText == nil {
		return nil
	}
	for y := 0; y < min(c.height, len(t.wrappedText)); y++ {
		for x, chr := range t.RenderLine(t.wrappedText[y], c.width) {
			tui.screen.SetContent(c.AbsoluteX()+x, c.AbsoluteY()+y, chr, nil, c.style)
		}
	}
	return nil
}

func (t *Text) OnUpdate(c *BaseComponent) error {
	if c.width > c.padding.Left+c.padding.Right { // cannot update if the width is 0 or negative TODO: protect the "calculateWordWrap" function better
		t.wrappedText = t.calculateWordWrap(c.width - c.padding.Left - c.padding.Right)
		c.fixedHeight = len(t.wrappedText) + c.padding.Top + c.padding.Bottom
	}
	return nil
}

func (t *Text) SetText(text string) {
	t.text = text
	t.wrappedText = nil
}

func (t *Text) String() string {
	align := ""
	switch t.alignment {
	case TextAlignLeft:
		align = "left"
	case TextAlignCenter:
		align = "center"
	case TextAlignRight:
		align = "right"
	case TextAlignJustify:
		align = "justify"
	}
	return fmt.Sprintf("Text (alignment: %s)", align)
}

func NewText(text string, layout Layout) BaseComponent {
	t := Text{text: text, alignment: TextAlignLeft}
	return NewComponent(&t, layout)
}
