package tui

import "fmt"

type mode string

const (
	normalMode mode = "normal"
	renameMode mode = "rename"
	newMode    mode = "new"

	normalModePrompt = "input > "
	renameModePrompt = "rename %s to > "
	newModePrompt    = "new session name > "
)

type modeState struct {
	prompt      string
	input       []rune
	placeholder *string
}

func (ms *modeState) getPrompt() string {
	if ms.placeholder == nil {
		return ms.prompt
	}

	return fmt.Sprintf(ms.prompt, *ms.placeholder)
}

func (ms *modeState) setInput(input []rune) {
	ms.input = input
}

func (ms *modeState) setPlaceholder(placeholder *string) {
	ms.placeholder = placeholder
}

func (ms *modeState) reset() {
	ms.input = []rune{}
	ms.placeholder = nil
}
