package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	// COLORS
	RED   = "\033[31m"
	GREEN = "\033[32m"
	BROWN = "\033[33m"
	CYAN  = "\033[36m"

	// TEXT EFFECTS
	RESET      = "\033[0m"
	BOLD       = "\033[1m"
	CLEAR_LINE = "\033[K"
)

type TUI struct {
	height    int
	sessions  []Session
	prompt    string
	directory string
	selected  int
	input     []rune
}

func NewTUI(height int, sessions []Session, prompt string, directory string) *TUI {
	return &TUI{
		height:    height,
		sessions:  sessions,
		prompt:    prompt,
		directory: directory,
	}
}

func (tui *TUI) SetPrompt(prompt string) {
	tui.prompt = prompt
}

func (tui *TUI) SetSessions(sessions []Session) {
	tui.sessions = sessions
}

func (tui *TUI) Render(query string) {
	fmt.Print(tui.generateFrame(query))
}

func (tui *TUI) Iterate(ke KeyEvent) {
	tui.handleKey(ke)
	inputString := string(tui.input)
	mode, query := parseInput(inputString)

	filteredSessions := fuzzyFilterSessions(tui.sessions, query)
	tui.adjustSelected(len(filteredSessions))
	tui.Render(query)

	if ke.SpecialKey == ENTER {
		var selectedSession *Session
		if len(filteredSessions) != 0 {
			selectedSession = &filteredSessions[tui.selected]
		}
		tui.handleCommand(mode, selectedSession, query)
		tui.Render(query)
	}
}

func (tui *TUI) generateFrame(query string) string {
	sessionRepresentations := fuzzyFilterSessionsColorized(tui.sessions, query)
	frame := "\033[H"
	frame += generateEmptyLines(tui.height - len(sessionRepresentations) - 2)
	frame += generateSessionsRepresentation(sessionRepresentations, tui.height, tui.selected)
	frame += generateFooter(len(sessionRepresentations), len(tui.sessions), tui.prompt, string(tui.input))
	return frame
}

func (tui *TUI) adjustSelected(sessionsCount int) {
	if tui.selected < 0 {
		tui.selected = 0
	} else if tui.selected >= sessionsCount {
		tui.selected = sessionsCount - 1
	}
}

func (tui *TUI) handleCommand(mode string, selectedSession *Session, query string) {
	if mode == "" || mode == ":rm" || mode == ":new" {
		tmux := os.Getenv("TMUX")
		if tmux == "" {
			os.Exit(1)
		}
	}

	switch mode {
	case "":
		sessionName := query
		if selectedSession != nil {
			sessionName = selectedSession.Name
		}

		if selectedSession == nil {
			createTmuxSession(sessionName, tui.directory)
		}
		switchTmuxSession(sessionName)
		os.Exit(0)
	case ":rm":
		if selectedSession != nil {
			sessionName := selectedSession.Name
			killTmuxSession(sessionName)

			newSessions := make([]Session, 0)
			for _, session := range tui.sessions {
				if session.Name != sessionName {
					newSessions = append(newSessions, session)
				}
			}
			tui.sessions = newSessions
		}
	case ":new":
		if query != "" {
			createTmuxSession(query, tui.directory)
			switchTmuxSession(query)
		}
		os.Exit(0)
	default:
		fmt.Println("Unknown command")
	}
}

func (tui *TUI) handleKey(ke KeyEvent) {
	switch ke.SpecialKey {
	case UP:
		tui.selected += 1
	case DOWN:
		tui.selected -= 1
	case BACKSPACE:
		if len(tui.input) > 0 {
			tui.input = tui.input[:len(tui.input)-1]
		}
	case ESC, EOT, ETX:
		os.Exit(0)
	case USUAL:
		tui.input = append(tui.input, ke.Key)
	}
}

func parseInput(input string) (string, string) {
	if input == "" {
		return "", ""
	}

	if input[0] == ':' {
		parts := strings.SplitN(input, " ", 2)
		if len(parts) == 1 {
			return parts[0], ""
		} else {
			return parts[0], parts[1]
		}
	}

	return "", input
}

func generateEmptyLines(count int) string {
	return strings.Repeat(CLEAR_LINE+"\r\n", count)
}

func generateSessionsRepresentation(representations []FuzzySearchResultRepresentation, height, selected int) string {
	frame := ""
	from := max(0, (selected+1)-(height-2))
	to := min(len(representations), from+height-2)
	for i := to - 1; i >= from; i-- {
		representation := representations[i]
		isSelected := i == selected

		if isSelected {
			frame += fmt.Sprintf("%s%s>%s%s", RED, BOLD, RESET, representation.GetString(isSelected))
		} else {
			frame += fmt.Sprintf(" %s", representation.GetString(isSelected))
		}

		frame += CLEAR_LINE + "\r\n"
	}
	return frame
}

func generateFooter(count, total int, prompt, input string) string {
	frame := BROWN + fmt.Sprintf("%d/%d", count, total) + RESET + CLEAR_LINE + "\r\n"
	frame += CYAN + prompt + input + RESET + CLEAR_LINE
	return frame
}
