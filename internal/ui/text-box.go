package ui

import (
	"hnterminal/tui"
)

type TextBox struct {
	root *tui.BaseComponent
	text *tui.BaseComponent
}

func NewTextBox(text string) TextBox {
	root := tui.NewBox(tui.FixedWidth)
	t := tui.NewText(text, tui.FixedWidth)
	textBox := TextBox{root: &root, text: &t}
	root.AddChild(textBox.text)

	return textBox
}

func (tb *TextBox) Text() *tui.BaseComponent {
	return tb.text
}

func (tb *TextBox) Root() *tui.BaseComponent {
	return tb.root
}

func (tb *TextBox) Padding() tui.Padding {
	return tb.root.Padding()
}

func (tb *TextBox) SetPadding(p tui.Padding) {
	tb.root.SetPadding(p)
}

func (tb *TextBox) SetBorderStyle(bs tui.BorderStyle) {
	tb.root.Kind().(*tui.Box).SetBorderStyle(bs)
}

func (tb *TextBox) SetBorder(b tui.Border) {
	tb.root.Kind().(*tui.Box).SetBorder(b)
}

func (tb *TextBox) SetText(t string) {
	tb.text.Kind().(*tui.Text).SetText(t)
}
