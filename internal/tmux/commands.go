package tmux

import (
	"github.com/verte-zerg/gession/internal/session"
	"github.com/verte-zerg/gession/pkg/assert"
	"strings"
)

type Command interface {
	GetCommand(escaping bool) string
	SetResult(content string)
}

// tmuxCommandCapturePane is a command to capture the content of a tmux pane.
type tmuxCommandCapturePane struct {
	PaneID   string
	Snapshot string
}

func (t tmuxCommandCapturePane) GetCommand(_ bool) string {
	return "capture-pane -p -e -t " + t.PaneID
}

func (t *tmuxCommandCapturePane) SetResult(result string) {
	t.Snapshot = result
}

// tmuxCommandListTree is a command to list all tmux sessions, windows, and panes.
type tmuxCommandListTree struct {
	Sessions []*session.Session
}

func (t tmuxCommandListTree) GetCommand(escaping bool) string {
	formatString := "#{session_name}|#{window_name}|#{pane_current_command}|#{window_index}.#{pane_index}|#{session_attached}.#{window_active}.#{pane_active}|#{session_last_attached}|#{session_id}.#{window_id}.#{pane_id}"
	if escaping {
		return "list-panes -a -F \"" + formatString + "\""
	}

	return "list-panes -a -F " + formatString
}

func (t *tmuxCommandListTree) SetResult(result string) {
	var err error
	t.Sessions, err = session.ParseSessions(result[:len(result)-1])

	if err != nil {
		panic(err)
	}
}

// tmuxCommandCurrentWindow is a command to get the current session and window ID.
type tmuxCommandCurrentWindow struct {
	SessionID string
	WindowID  string
}

func (t tmuxCommandCurrentWindow) GetCommand(escaping bool) string {
	if escaping {
		return "display-message -p -F \"#{session_id}.#{window_id}\""
	}

	return "display-message -p -F #{session_id}.#{window_id}"
}

func (t *tmuxCommandCurrentWindow) SetResult(result string) {
	ids := strings.Split(result[:len(result)-1], ".")
	assert.Assert(len(ids) == 2, "Invalid format of fetched session and window ID") //nolint:mnd
	t.SessionID, t.WindowID = ids[0], ids[1]
}
