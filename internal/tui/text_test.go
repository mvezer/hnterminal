package tui

import (
	"testing"
)

func testWordWrap(text *BaseComponent, expected [][]string, t *testing.T) {
	actual := text.kind.(*Text).wrappedText
	if len(actual) != len(expected) {
		t.Errorf("Expected %d lines, got %d", len(expected), len(actual))
	}
	for i := range actual {
		if len(actual[i]) != len(expected[i]) {
			t.Errorf("Expected %d words in line %d, got %d", len(expected[i]), i, len(actual[i]))
		}
		for j := range actual[i] {
			if actual[i][j] != expected[i][j] {
				t.Errorf("Expected word %d in line %d to be %s, got %s", j, i, expected[i][j], actual[i][j])
			}
		}
	}
}

func Test_Text_WordWrapBasicCase(t *testing.T) {
	text := NewText(
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.", HorizontalGrid,
	)
	text.SetGeometry(0, 0, 25, 10)

	testWordWrap(
		&text, [][]string{
			{"Lorem", "ipsum", "dolor", "sit"},
			{"amet,", "consectetur"},
			{"adipiscing", "elit,", "sed", "do"},
			{"eiusmod", "tempor", "incididunt"},
			{"ut", "labore", "et", "dolore", "magna"},
			{"aliqua."},
		}, t)
}

func Test_Text_WordWrapWithNewlines(t *testing.T) {
	text := NewText(
		"Lorem\nipsum dolor sit amet, cons\nect\netur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.", HorizontalGrid,
	)
	text.SetGeometry(0, 0, 25, 10)
	testWordWrap(&text, [][]string{
		{"Lorem"},
		{"ipsum", "dolor", "sit", "amet,"},
		{"cons"},
		{"ect"},
		{"etur", "adipiscing", "elit,", "sed"},
		{"do", "eiusmod", "tempor"},
		{"incididunt", "ut", "labore", "et"},
		{"dolore", "magna", "aliqua."},
	}, t)
}

func Test_Text_WordWrapWithLongWords(t *testing.T) {
	text := NewText(
		"Loremipsumdolor sit amet, consectetur adipiscing elit, seddoeiusmodtemporincididuntutlabore et dolore magna aliqua.", HorizontalGrid,
	)
	text.SetGeometry(0, 0, 10, 10)
	testWordWrap(&text, [][]string{
		{"Loremipsum"},
		{"dolor", "sit"},
		{"amet,", "conse"},
		{"ctetur"},
		{"adipiscing"},
		{"elit,", "seddo"},
		{"eiusmodtem"},
		{"porincidid"},
		{"untutlabor"},
		{"e", "et", "dolore"},
		{"magna"},
		{"aliqua."},
	}, t)
}

func testRenderLine(text *BaseComponent, expected string, t *testing.T) {
	actual := text.kind.(*Text).RenderLine(text.kind.(*Text).wrappedText[0], text.width)
	if actual != expected {
		t.Errorf("Expected \"%s\", got \"%s\"", expected, actual)
	}
}

func Test_Text_RenderLineLeftAligned(t *testing.T) {
	text := NewText("One two three", HorizontalGrid)
	text.kind.(*Text).SetAlignment(TextAlignLeft)
	text.SetGeometry(0, 0, 20, 10)
	testRenderLine(&text, "One two three", t)
}

func Test_Text_RenderLineCenterAligned(t *testing.T) {
	text := NewText("One two three", HorizontalGrid)
	text.kind.(*Text).SetAlignment(TextAlignCenter)
	text.SetGeometry(0, 0, 20, 10)
	testRenderLine(
		&text, "   One two three", t)
}

func Test_Text_RenderLineRightAligned(t *testing.T) {
	text := NewText("One two three", HorizontalGrid)
	text.kind.(*Text).SetAlignment(TextAlignRight)
	text.SetGeometry(0, 0, 20, 10)
	testRenderLine(&text, "       One two three", t)
}

func Test_Text_RenderLineJustfifyAligned(t *testing.T) {
	text := NewText("One two three", HorizontalGrid)
	text.kind.(*Text).SetAlignment(TextAlignJustify)
	text.SetGeometry(0, 0, 20, 10)
	testRenderLine(&text, "One     two    three", t)
}
