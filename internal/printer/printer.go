package printer

import (
	"fmt"
	"github.com/verte-zerg/gession/internal/sessiontree"
	"github.com/verte-zerg/gession/pkg/ansi"
	"github.com/verte-zerg/gession/pkg/assert"
	"slices"
	"strings"
	"unicode/utf8"
)

const (
	// COLORS.
	cursor            = "\033[31m"
	highlight         = "\033[32m"
	sessionStats      = "\033[38;5;180m"
	prompt            = "\033[38;5;111m"
	hotkeyDescription = "\033[38;5;240m"
	hotkeyKey         = "\033[38;5;239m"
	hotkeySeparator   = "\033[38;5;238m"

	// TEXT EFFECTS.
	reset     = "\033[0m"
	bold      = "\033[1m"
	clearLine = "\033[K"

	// CURSOR.
	hideCursor = "\033[?25l"
	showCursor = "\033[?25h"
	jumpCell   = "\033[%dA\033[%dG"

	// TUI.
	footerHeight = 3
)

var (
	footerHotkeys = [][2]string{
		{"<c-c/d>", "exit"},
		{"<c-e>", "delete"},
		{"<c-r>", "rename"},
		{"<c-t>", "new"},
		{"←/→", "wrap/unwrap"},
		{"↑/↓/tab/<s-tab>", "move"},
		{"enter", "select/create"},
	}
	footerPrimeHotkeys = [][2]string{
		{"<c-c/d>", "exit"},
		{"↑/↓/tab/<s-tab>", "move"},
		{"enter", "select/create"},
	}
	normalFooter = newFooter(footerHotkeys)
	primeFooter  = newFooter(footerPrimeHotkeys)
)

type footer struct {
	linesBySize []footerLine
}

func (f footer) String(width int) string {
	for _, line := range f.linesBySize {
		if line.size <= width {
			return line.line
		}
	}

	assert.Fatal("no footer line fits the width")
	panic("unreachable")
}

type footerLine struct {
	line string
	size int
}

func newFooter(hotkeys [][2]string) *footer {
	line := ""

	footerLines := make([]footerLine, 0)

	footerLines = append(footerLines, footerLine{"", 0})

	for i, hotkey := range hotkeys {
		line += fmt.Sprintf("%[1]s%[2]s %[3]s%[4]s", hotkeyDescription, hotkey[0], hotkeyKey, hotkey[1])

		lineLen := ansi.CalculateVisibleLen(line)
		footerLines = append(footerLines, footerLine{line, lineLen})

		if i < len(hotkeys)-1 {
			line += hotkeySeparator + " • "
		}
	}

	slices.Reverse(footerLines)

	return &footer{footerLines}
}

type Printer struct {
	width  int
	height int

	prime bool
}

func New(width, height int, prime bool) *Printer {
	return &Printer{width, height, prime}
}

func (p Printer) GenerateFrame(vTree *sessiontree.VisualizeTree, input string) string {
	frame := "\033[H"

	selectedSession := vTree.GetSelectedSession()
	selectedWindow := vTree.GetSelectedWindow()
	filteredSessionsCount := len(vTree.GetSessions())
	rows := vTree.GetVisibleRows()

	restHeight := p.height

	if selectedSession != nil && !p.prime {
		previewHeight := p.height / 2 //nolint:mnd
		restHeight = p.height - previewHeight

		var windowID *string

		if selectedWindow != nil {
			windowID = &selectedWindow.ID
		}

		frame += p.generateSessionPreview(*selectedSession, windowID, previewHeight, p.width)
	}

	frame += p.generateEmptyLines(restHeight - rows - footerHeight)
	frame += p.generateSessionsRepresentation(vTree, restHeight)
	frame += p.generateFooter(filteredSessionsCount, vTree.GetSessionsCount(), input)

	return hideCursor + frame + showCursor
}

func (p Printer) generateSessionPreview(session sessiontree.FilteredSession, windowID *string, height, width int) string {
	panesSnapshots := make([]*string, 0)

	for _, window := range session.FilteredChildren {
		if windowID == nil || window.ID == *windowID {
			for _, pane := range window.FilteredChildren {
				panesSnapshots = append(panesSnapshots, pane.Snapshot)
			}
		}
	}

	separatorsCount := len(panesSnapshots) + 1
	contentWidth := width - separatorsCount
	contentHeight := height - footerHeight
	widths := make([]int, len(panesSnapshots))
	defaultWidth := contentWidth / len(panesSnapshots)
	restWidth := contentWidth % len(panesSnapshots)

	for i := range widths {
		widths[i] = defaultWidth
		if i < restWidth {
			widths[i]++
		}
	}

	snapshotsLines := make([][]ansi.Line, 0)

	for i, snapshot := range panesSnapshots {
		cuttedSnapshotLines := ansi.CutSnapshot(snapshot, widths[i], contentHeight)
		snapshotsLines = append(snapshotsLines, cuttedSnapshotLines)
	}

	DEL := reset + "│"

	lines := make([]string, contentHeight+footerHeight)

	for i := range contentHeight {
		for j, cuttedSnapshotLines := range snapshotsLines {
			if i < len(cuttedSnapshotLines) {
				lines[i+1] += DEL + cuttedSnapshotLines[i].Content + strings.Repeat(" ", widths[j]-cuttedSnapshotLines[i].Len)
			} else {
				lines[i+1] += DEL + strings.Repeat(" ", widths[j])
			}
		}

		lines[i+1] += DEL
	}

	lineDashParts := make([]string, 0)

	for _, width := range widths {
		lineDashParts = append(lineDashParts, strings.Repeat("─", width))
	}

	lines[0] = reset + "┌" + strings.Join(lineDashParts, "┬") + "┐"
	lines[contentHeight+1] = reset + "└" + strings.Join(lineDashParts, "┴") + "┘"

	return strings.Join(lines, clearLine+"\r\n") + "\r\n"
}

func (p Printer) generateEmptyLines(count int) string {
	if count <= 0 {
		return ""
	}

	return strings.Repeat(clearLine+"\r\n", count)
}

func (p Printer) generateWindowsRepresentation(
	windows []*sessiontree.FilteredWindow,
	orderIdx int,
	displayFrom int,
	displayTo int,
	selected int,
) []string {
	lines := make([]string, 0)

	for windowOrderIdx, window := range windows {
		if orderIdx >= displayFrom && orderIdx < displayTo {
			var line string

			subTreeChar := "├─"

			// if it's last window in session
			if windowOrderIdx == 0 {
				subTreeChar = "└─"
			}

			if orderIdx == selected {
				line = fmt.Sprintf("%s%s> %s%s %s", cursor, bold, reset, subTreeChar, window.GetString(true))
			} else {
				line = fmt.Sprintf("  %s %s", subTreeChar, window.GetString(false))
			}

			if window.IsActive {
				line += " (active)"
			}

			line += clearLine + "\r\n"
			lines = append(lines, line)
		}

		orderIdx++
	}

	return lines
}

func (p Printer) generateSessionRepresentation(
	session *sessiontree.FilteredSession,
	isSelected bool,
) string {
	var line, unwrapChar string

	switch {
	case session.IsUnwrapped && !p.prime:
		unwrapChar = "+ "
	case !session.IsUnwrapped && !p.prime:
		unwrapChar = "- "
	}

	if isSelected {
		line = fmt.Sprintf("%s%s>%s%s %s%s%s", cursor, bold, reset, bold, unwrapChar, reset, session.GetString(true))
	} else {
		line = fmt.Sprintf("  %s%s", unwrapChar, session.GetString(false))
	}

	if session.IsAttached {
		line += " (attached)"
	}

	if p.prime && !strings.HasPrefix(session.ID, "notexisted_") {
		line += " (existed)"
	}

	line += clearLine + "\r\n"

	return line
}

func (p Printer) generateSessionsRepresentation(vTree *sessiontree.VisualizeTree, height int) string {
	sessions := vTree.GetSessions()
	selected := vTree.GetSelectedIdx()
	visibleRows := vTree.GetVisibleRows()

	displayFrom := max(0, (selected+1)-(height-footerHeight))
	dispayTo := min(visibleRows, displayFrom+height-footerHeight)

	lines := make([]string, 0)

	orderIdx := 0

	for _, session := range sessions {
		if session.IsUnwrapped {
			windowsLines := p.generateWindowsRepresentation(session.FilteredChildren, orderIdx, displayFrom, dispayTo, selected)
			lines = append(lines, windowsLines...)
			orderIdx += len(session.FilteredChildren)
		}

		if orderIdx >= displayFrom && orderIdx < dispayTo {
			isSelected := orderIdx == selected
			sessionLine := p.generateSessionRepresentation(&session, isSelected)
			lines = append(lines, sessionLine)
		}

		orderIdx++
	}

	slices.Reverse(lines)

	return strings.Join(lines, "")
}

func (p Printer) generateFooter(count, total int, input string) string {
	frame := prompt + input + reset + clearLine + "\r\n"
	frame += sessionStats + fmt.Sprintf("sessions: %d/%d", count, total) + reset + clearLine + "\r\n"
	hotkeyList := normalFooter.String(p.width)

	if p.prime {
		hotkeyList = primeFooter.String(p.width)
	}

	frame += hotkeyDescription + hotkeyList + clearLine + reset + relativelyJumpToCell(footerHeight-1, utf8.RuneCountInString(input)+1)

	return frame
}

func relativelyJumpToCell(row, col int) string {
	return fmt.Sprintf(jumpCell, row, col)
}
