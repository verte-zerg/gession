package tmux

import (
	"github.com/verte-zerg/gession/pkg/assert"
	"os/exec"
)

func CreateTmuxSession(name string, directory string) {
	tmux := exec.Command("tmux", "new-session", "-d", "-s", name, "-c", directory)
	err := tmux.Run()
	assert.Assert(err == nil, "Failed to create tmux session")
}

func CreateTmuxWindow(sessionName, name, directory string) {
	tmux := exec.Command("tmux", "new-window", "-d", "-n", name, "-c", directory, "-t", sessionName+":")
	err := tmux.Run()
	assert.Assert(err == nil, "Failed to create tmux window")
}

func SwitchClient(entityID string) {
	tmux := exec.Command("tmux", "switch-client", "-t", entityID)
	err := tmux.Run()
	assert.Assert(err == nil, "Failed to switch tmux session")
}

func KillTmuxSession(sessionID string) {
	tmux := exec.Command("tmux", "kill-session", "-t", sessionID)
	err := tmux.Run()
	assert.Assert(err == nil, "Failed to kill tmux session")
}

func KillTmuxWindow(windowID string) {
	tmux := exec.Command("tmux", "kill-window", "-t", windowID)
	err := tmux.Run()
	assert.Assert(err == nil, "Failed to kill tmux window")
}

func RenameTmuxSession(sessionID, newName string) {
	tmux := exec.Command("tmux", "rename-session", "-t", sessionID, newName)
	err := tmux.Run()
	assert.Assert(err == nil, "Failed to rename tmux session")
}

func RenameTmuxWindow(windowID, newName string) {
	tmux := exec.Command("tmux", "rename-window", "-t", windowID, newName)
	err := tmux.Run()
	assert.Assert(err == nil, "Failed to rename tmux window")
}
