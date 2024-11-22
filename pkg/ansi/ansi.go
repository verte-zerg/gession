package ansi

import (
	"regexp"
	"strings"

	textwidth "golang.org/x/text/width"
)

// RegEX to match ANSI escape sequences.
var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*[mGKHF]`)

const (
	NormalWidth = 1
	WideWidth   = 2
)

type Line struct {
	Content string
	Len     int
}

func CutSnapshot(snapshot *string, width, height int) []Line {
	s := ""
	if snapshot != nil {
		s = *snapshot
	}

	lines := strings.Split(s, "\n")
	cuttedLines := make([]Line, 0)

	if len(lines) > height {
		lines = lines[:height]
	}

	ansiState := ""

	for lineIdx := range lines {
		style, fgColor, bgColor := DetectEscapeSequence(ansiState + lines[lineIdx])
		codes := make([]string, 0)
		codes = append(codes, style.GetCode())

		if fgColor != nil {
			codes = append(codes, fgColor.GetCode(true))
		}

		if bgColor != nil {
			codes = append(codes, bgColor.GetCode(false))
		}

		ansiState = "\x1b[" + strings.Join(codes, ";") + "m"
		cuttedLines = append(cuttedLines, CutString(ansiState+lines[lineIdx], width))
	}

	return cuttedLines
}

func getRuneVisibleLen(r rune) int {
	if k := textwidth.LookupRune(r).Kind(); k == textwidth.EastAsianWide || k == textwidth.EastAsianFullwidth {
		return WideWidth
	}

	return NormalWidth
}

// Calculate visible length of a string.
func CalculateVisibleLen(str string) int {
	visibleLen := 0
	pos := 0

	matches := ansiRegexp.FindAllStringIndex(str, -1)

	for pos < len(str) {
		if len(matches) > 0 && matches[0][0] == pos {
			pos = matches[0][1]
			matches = matches[1:]
		} else {
			for _, r := range str[pos:] {
				runeLen := getRuneVisibleLen(r)
				visibleLen += runeLen
				pos += len(string(r))

				break
			}
		}
	}

	return visibleLen
}

// Function to cut an ANSI string to a fixed width.
func CutString(str string, width int) Line {
	var result strings.Builder

	visibleLen := 0
	pos := 0

	matches := ansiRegexp.FindAllStringIndex(str, -1)

	for pos < len(str) {
		if len(matches) > 0 && matches[0][0] == pos {
			result.WriteString(str[pos:matches[0][1]])
			pos = matches[0][1]
			matches = matches[1:]
		} else {
			for _, r := range str[pos:] {
				runeLen := getRuneVisibleLen(r)

				if visibleLen+runeLen <= width {
					visibleLen += runeLen

					result.WriteRune(r)
				}

				pos += len(string(r))

				break
			}
		}

		if visibleLen >= width {
			break
		}
	}

	return Line{Content: result.String(), Len: visibleLen}
}
