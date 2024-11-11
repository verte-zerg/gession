package tui

import (
	"github.com/verte-zerg/gession/internal/event"
	"github.com/verte-zerg/gession/internal/key"
	"log/slog"
	"os"
)

func (tui *TUI) deleteChar() {
	input := tui.modeStates[tui.mode].input
	if len(input) == 0 {
		return
	}

	input = input[:len(input)-1]

	ms := tui.modeStates[tui.mode]
	ms.setInput(input)

	logger.Info("input cutted", slog.String("input", string(ms.input)), slog.String("mode", string(tui.mode)))
}

func (tui *TUI) addChar(char rune) {
	ms := tui.modeStates[tui.mode]
	ms.setInput(append(ms.input, char))

	logger.Info("input appended", slog.String("input", string(ms.input)), slog.String("mode", string(tui.mode)))
}

func (tui *TUI) handleMoving(keyEvent event.KeyPressed) (refilteringRequired bool) {
	if tui.mode != normalMode {
		return false
	}

	//nolint:exhaustive
	switch keyEvent.SpecialKey {
	case key.Up, key.ShiftTab:
		tui.selectedIdx++
		logger.Info("Moved selection up", slog.Int("selectedIdx", tui.selectedIdx))

		return true
	case key.Down, key.Tab:
		tui.selectedIdx--
		logger.Info("Moved selection down", slog.Int("selectedIdx", tui.selectedIdx))

		return true
	case key.Left:
		if tui.kind == PrimeKind {
			return false
		}

		selectedSession := tui.vTree.GetSelectedSession()
		if selectedSession == nil {
			return false
		}

		if _, ok := tui.unwrappedSession[selectedSession.ID]; ok {
			tui.selectedIdx = tui.vTree.GetFirstSelectedWindowIdx()
			delete(tui.unwrappedSession, selectedSession.ID)
			logger.Info("Unwrapped session", slog.String("sessionID", selectedSession.ID))
		} else {
			tui.selectedIdx++
		}

		return true
	case key.Right:
		if tui.kind == PrimeKind {
			return false
		}

		selectedSession := tui.vTree.GetSelectedSession()
		if selectedSession == nil {
			return false
		}

		if _, ok := tui.unwrappedSession[selectedSession.ID]; !ok {
			tui.unwrappedSession[selectedSession.ID] = struct{}{}
			tui.selectedIdx += len(selectedSession.FilteredChildren)
			logger.Info("Wrapped session", slog.String("sessionID", selectedSession.ID))
		} else {
			tui.selectedIdx--
		}

		return true
	}

	return false
}

//nolint:gocritic,cyclop,funlen
func (tui *TUI) handleKeyEvent(keyEvent event.KeyPressed) {
	logger.Info("key event", slog.String("key", string(keyEvent.Key)), slog.String("specialKey", string(keyEvent.SpecialKey)))

	if keyEvent.SpecialKey == key.Ignore {
		return
	}

	defer tui.Render()

	refilteringRequired := false

	defer func() {
		if refilteringRequired {
			tui.filterSessions()
		}
	}()

	//nolint:exhaustive
	switch keyEvent.SpecialKey {
	// Exit on Ctrl+D, Ctrl+C
	case key.EOT, key.ETX:
		logger.Info("Exiting application")
		os.Exit(0)

	// Reset mode to NORMAL, in NORMAL mode exit on Esc
	case key.Esc:
		if tui.mode == normalMode {
			logger.Info("Exiting application")
			os.Exit(0)
		} else {
			ms := tui.modeStates[tui.mode]
			ms.reset()

			tui.mode = normalMode
		}

	// Delete last char
	case key.Backspace:
		tui.deleteChar()

		if tui.mode == normalMode {
			refilteringRequired = true
		}

	// Enter to NEW mode
	case key.CtrlT:
		if tui.kind == NormalKind {
			tui.mode = newMode

			logger.Info("Switched to new mode")
		}

	// Add char to input
	case key.Usual:
		tui.addChar(keyEvent.Key)

		if tui.mode == normalMode {
			refilteringRequired = true
		}

	// Moving in menu
	case key.Up, key.Down, key.Left, key.Right, key.Tab, key.ShiftTab:
		if tui.handleMoving(keyEvent) {
			refilteringRequired = true
		}

	// Run command depending on mode
	case key.Enter:
		ms := tui.modeStates[tui.mode]
		tui.handleCommand(string(ms.input), false)

		if tui.mode == renameMode {
			ms.reset()

			refilteringRequired = true
			tui.mode = normalMode
		}

	// RENAME mode
	case key.CtrlR:
		if tui.kind == PrimeKind {
			return
		}

		selectedSession := tui.vTree.GetSelectedSession()

		if selectedSession == nil {
			return
		}

		tui.mode = renameMode
		ms := tui.modeStates[renameMode]

		placeholder := selectedSession.Name
		if selectedWindow := tui.vTree.GetSelectedWindow(); selectedWindow != nil {
			placeholder += ":" + selectedWindow.Name
		}

		ms.setPlaceholder(&placeholder)

	// Delete session or window
	case key.CtrlE:
		if tui.kind == PrimeKind {
			return
		}

		selectedSession := tui.vTree.GetSelectedSession()
		if selectedSession == nil {
			return
		}

		tui.handleCommand("", true)
		refilteringRequired = true
	}
}
