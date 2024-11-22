package tmux

import (
	"github.com/verte-zerg/gession/internal/event"
	"github.com/verte-zerg/gession/pkg/assert"
)

func ConvertEventToCommand(e event.Event) Command {
	switch e.Type {
	case event.TypeCapturePane:
		eventPane, ok := e.Data.(event.CapturePane)
		assert.Assert(ok, "event.Data is not a EventCapturePane")

		return &tmuxCommandCapturePane{
			PaneID: eventPane.PaneID,
		}
	case event.TypeListTree:
		return &tmuxCommandListTree{}
	default:
		assert.Fatal("Unknown event type")
	}

	panic("Unreachable")
}

func ConvertCommandToEvent(command Command) event.Event {
	if command, ok := command.(*tmuxCommandCapturePane); ok {
		return event.Event{
			Type: event.TypeCapturedPane,
			Data: event.CapturedPane{
				PaneID:   command.PaneID,
				Snapshot: command.Snapshot,
			},
		}
	}

	if command, ok := command.(*tmuxCommandListTree); ok {
		return event.Event{
			Type: event.TypeListedTree,
			Data: event.ListedTree{
				Sessions: command.Sessions,
			},
		}
	}

	assert.Fatal("Unknown command type")
	panic("Unreachable")
}
