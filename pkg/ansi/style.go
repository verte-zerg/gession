package ansi

import (
	"fmt"
	"regexp"
	"strings"
)

var ansiRegexpColor = regexp.MustCompile(`\x1b\[[0-9;]*m`)

type Style struct {
	Bold, Dim, Italic, Underline, Blink, Inverse, Hidden, Strikethrough bool
}

//nolint:cyclop
func (s *Style) HandleEscapeSequence(code string) {
	switch code {
	case "0":
		*s = Style{}
	case "1":
		s.Bold = true
	case "2":
		s.Dim = true
	case "3":
		s.Italic = true
	case "4":
		s.Underline = true
	case "5":
		s.Blink = true
	case "7":
		s.Inverse = true
	case "8":
		s.Hidden = true
	case "9":
		s.Strikethrough = true
	case "22":
		s.Bold = false
		s.Dim = false
	case "23":
		s.Italic = false
	case "24":
		s.Underline = false
	case "25":
		s.Blink = false
	case "27":
		s.Inverse = false
	case "28":
		s.Hidden = false
	case "29":
		s.Strikethrough = false
	}
}

func (s Style) GetCode() string {
	modes := make([]string, 0)

	if s.Bold {
		modes = append(modes, "1")
	}

	if s.Dim {
		modes = append(modes, "2")
	}

	if s.Italic {
		modes = append(modes, "3")
	}

	if s.Underline {
		modes = append(modes, "4")
	}

	if s.Blink {
		modes = append(modes, "5")
	}

	if s.Inverse {
		modes = append(modes, "7")
	}

	if s.Hidden {
		modes = append(modes, "8")
	}

	if s.Strikethrough {
		modes = append(modes, "9")
	}

	return strings.Join(modes, ";")
}

type Color interface {
	GetCode(fg bool) string
}

type Color16 struct {
	Code string
}

func newColor16(s string) (Color16, bool) {
	switch s {
	case "30", "31", "32", "33", "34", "35", "36", "37", "39", "90", "91", "92", "93", "94", "95", "96", "97":
		return Color16{Code: s}, true
	case "40", "41", "42", "43", "44", "45", "46", "47", "49", "100", "101", "102", "103", "104", "105", "106", "107":
		return Color16{Code: s}, false
	default:
		return Color16{Code: ""}, false
	}
}

func (c Color16) GetCode(_ bool) string {
	return c.Code
}

type Color256 struct {
	Code string
}

func (c Color256) GetCode(fg bool) string {
	if fg {
		return "38;5;" + c.Code
	}

	return "48;5;" + c.Code
}

type ColorRGB struct {
	R, G, B string
}

func (c ColorRGB) GetCode(fg bool) string {
	if fg {
		return fmt.Sprintf("38;2;%s;%s;%s", c.R, c.G, c.B)
	}

	return fmt.Sprintf("48;2;%s;%s;%s", c.R, c.G, c.B)
}

//nolint:gocognit,cyclop
func DetectEscapeSequence(str string) (Style, Color, Color) {
	matches := ansiRegexpColor.FindAllStringIndex(str, -1)
	if len(matches) == 0 {
		return Style{}, nil, nil
	}

	style := Style{}

	var fgColor, bgColor Color

	for _, match := range matches {
		codes := strings.Split(str[match[0]+2:match[1]-1], ";")
		isExtFgColor := false
		isExtBgColor := false
		isColor256 := false
		isColorRGB := false

		for codeIdx := 0; codeIdx < len(codes); codeIdx++ {
			code := codes[codeIdx]
			if !isExtBgColor && !isExtFgColor {
				if code == "38" {
					isExtFgColor = true

					continue
				}

				if code == "48" {
					isExtBgColor = true

					continue
				}
			}

			if (isExtFgColor || isExtBgColor) && (!isColor256 && !isColorRGB) {
				if code == "5" {
					isColor256 = true

					continue
				}

				if code == "2" {
					isColorRGB = true

					continue
				}

				isExtFgColor = false
				isExtBgColor = false
			}

			if isColor256 || isColorRGB {
				var clr Color

				if isColor256 {
					clr = &Color256{Code: code}
				}

				if isColorRGB && codeIdx+2 < len(codes) {
					clr = &ColorRGB{R: codes[codeIdx], G: codes[codeIdx+1], B: codes[codeIdx+2]}
					codeIdx += 2
				}

				if isExtFgColor && clr != nil {
					fgColor = clr
				}

				if isExtBgColor && clr != nil {
					bgColor = clr
				}

				isExtFgColor = false
				isExtBgColor = false
				isColor256 = false
				isColorRGB = false

				continue
			}

			c16, isFg := newColor16(code)
			if c16.Code != "" {
				if isFg {
					fgColor = c16
				} else {
					bgColor = c16
				}

				continue
			}

			style.HandleEscapeSequence(code)

			if code == "0" {
				fgColor = nil
				bgColor = nil
			}
		}
	}

	return style, fgColor, bgColor
}
