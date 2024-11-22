package ansi_test

import (
	"testing"

	"github.com/verte-zerg/gession/pkg/ansi"
)

func TestDetectAnsiEscapeSequence_AllCases(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedStyle ansi.Style
		expectedFg    ansi.Color
		expectedBg    ansi.Color
	}{
		// Empty string
		{
			name:          "Empty string",
			input:         "",
			expectedStyle: ansi.Style{},
			expectedFg:    nil,
			expectedBg:    nil,
		},
		// Basic style (bold, italic, underline)
		{
			name:          "Basic style",
			input:         "\x1b[1;3;4m",
			expectedStyle: ansi.Style{Bold: true, Italic: true, Underline: true},
			expectedFg:    nil,
			expectedBg:    nil,
		},
		// 16-color foreground and background
		{
			name:          "16-color foreground and background",
			input:         "\x1b[31;47m",
			expectedStyle: ansi.Style{},
			expectedFg:    ansi.Color16{Code: "31"},
			expectedBg:    ansi.Color16{Code: "47"},
		},
		// 256-color foreground and background
		{
			name:          "256-color foreground and background",
			input:         "\x1b[38;5;123;48;5;234m",
			expectedStyle: ansi.Style{},
			expectedFg:    ansi.Color256{Code: "123"},
			expectedBg:    ansi.Color256{Code: "234"},
		},
		// RGB color foreground and background
		{
			name:          "RGB color foreground and background",
			input:         "\x1b[38;2;255;100;50;48;2;10;20;30m",
			expectedStyle: ansi.Style{},
			expectedFg:    ansi.ColorRGB{R: "255", G: "100", B: "50"},
			expectedBg:    ansi.ColorRGB{R: "10", G: "20", B: "30"},
		},
		// Mixed (bold, 256-color foreground, RGB background)
		{
			name:          "Mixed",
			input:         "\x1b[1;38;5;123;48;2;255;255;0m",
			expectedStyle: ansi.Style{Bold: true},
			expectedFg:    ansi.Color256{Code: "123"},
			expectedBg:    ansi.ColorRGB{R: "255", G: "255", B: "0"},
		},
		// Multiple sequences (bold, reset, italic)
		{
			name:          "Multiple sequences",
			input:         "\x1b[1m\x1b[22m\x1b[3m",
			expectedStyle: ansi.Style{Italic: true},
			expectedFg:    nil,
			expectedBg:    nil,
		},
		// Multiple sequences with color and reset
		{
			name:          "Multiple sequences with color and reset",
			input:         "\x1b[1m\x1b[31m\x1b[22m\x1b[39m",
			expectedStyle: ansi.Style{},
			expectedFg:    nil,
			expectedBg:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			style, fg, bg := ansi.DetectEscapeSequence(test.input)

			if style != test.expectedStyle {
				t.Errorf("For input %q, expected style %+v, got %+v", test.input, test.expectedStyle, style)
			}

			if (fg != nil && test.expectedFg != nil && fg.GetCode(true) != test.expectedFg.GetCode(true)) || (fg == nil && test.expectedFg != nil) {
				t.Errorf("For input %q, expected fg color %+v, got %+v", test.input, test.expectedFg, fg)
			}

			if (bg != nil && test.expectedBg != nil && bg.GetCode(false) != test.expectedBg.GetCode(false)) || (bg == nil && test.expectedBg != nil) {
				t.Errorf("For input %q, expected bg color %+v, got %+v", test.input, test.expectedBg, bg)
			}
		})
	}
}
