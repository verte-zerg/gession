package ansi_test

import (
	"github.com/verte-zerg/gession/pkg/ansi"
	"testing"
)

func TestVisibleLen(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Simple ASCII string",
			input:    "Hello, World!",
			expected: 13,
		},
		{
			name:     "ANSI escape sequence string",
			input:    "\x1b[31mHello, World!\x1b[0m",
			expected: 13,
		},
		{
			name:     "ANSI string with multibyte characters",
			input:    "\x1b[32m你好, 世界\x1b[0m",
			expected: 10,
		},
		{
			name:     "Only ANSI sequences, no visible characters",
			input:    "\x1b[31m\x1b[0m",
			expected: 0,
		},
		{
			name:     "Beer emoji",
			input:    "🍺",
			expected: 2,
		},
		{
			name:     "Beer emoji and ASCII",
			input:    "🍺Hello, World!",
			expected: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ansi.CalculateVisibleLen(tt.input)
			if result != tt.expected {
				t.Errorf("Expected length to be `%d`, got `%d` for input `%s`", tt.expected, result, tt.input)
			}
		})
	}
}

func TestCutString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected ansi.Line
	}{
		{
			name:     "Simple ASCII string",
			input:    "Hello, World!",
			width:    5,
			expected: ansi.Line{Content: "Hello", Len: 5},
		},
		{
			name:     "ANSI escape sequence string",
			input:    "\x1b[31mHello, World!\x1b[0m",
			width:    5,
			expected: ansi.Line{Content: "\x1b[31mHello", Len: 5},
		},
		{
			name:     "Shorter string than width",
			input:    "Hi",
			width:    10,
			expected: ansi.Line{Content: "Hi", Len: 2},
		},
		{
			name:     "Empty string",
			input:    "",
			width:    5,
			expected: ansi.Line{Content: "", Len: 0},
		},
		{
			name:     "ANSI string with multibyte characters",
			input:    "\x1b[32m你好, 世界\x1b[0m",
			width:    4,
			expected: ansi.Line{Content: "\x1b[32m你好", Len: 4},
		},
		{
			name:     "Only ANSI sequences, no visible characters",
			input:    "\x1b[31m\x1b[0m",
			width:    5,
			expected: ansi.Line{Content: "\x1b[31m\x1b[0m", Len: 0},
		},
		{
			name:     "Beer emoji",
			input:    "🍺",
			width:    2,
			expected: ansi.Line{Content: "🍺", Len: 2},
		},
		{
			name:     "Beer emoji cut off",
			input:    "🍺",
			width:    1,
			expected: ansi.Line{Content: "", Len: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ansi.CutString(tt.input, tt.width)
			if result.Content != tt.expected.Content {
				t.Errorf("Expected line to be `%s`, got `%s` for input `%s`", tt.expected.Content, result.Content, tt.input)
			}
		})
	}
}
