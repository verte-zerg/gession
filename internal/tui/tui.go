package tui

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/verte-zerg/gession/internal/event"
	"github.com/verte-zerg/gession/internal/printer"
	"github.com/verte-zerg/gession/internal/session"
	"github.com/verte-zerg/gession/internal/sessiontree"
	"github.com/verte-zerg/gession/internal/tmux"
	"github.com/verte-zerg/gession/pkg/assert"
	"github.com/verte-zerg/gession/pkg/logging"
)

type Kind int

const (
	NormalKind Kind = iota
	PrimeKind
)

var (
	logger = logging.GetInstance().WithGroup("tui")
)

type TUI struct {
	kind Kind

	mode       mode
	modeStates map[mode]*modeState

	sessions    []*session.Session
	tmpSessions []*session.Session
	directory   string

	printer *printer.Printer
	vTree   *sessiontree.VisualizeTree

	sessionIDToSession map[string]*session.Session
	paneIDToSession    map[string]*session.Session

	selectedIdx int

	unwrappedSession map[string]interface{}

	eventInputCh  chan event.Event
	eventOutputCh chan event.Event
}

func NewTUI(width, height int, kind Kind, directory string) *TUI {
	isPrimeKind := kind == PrimeKind

	return &TUI{
		kind:             kind,
		sessions:         make([]*session.Session, 0),
		tmpSessions:      nil,
		directory:        directory,
		eventInputCh:     make(chan event.Event, event.MaxQueue),
		unwrappedSession: make(map[string]interface{}),
		printer:          printer.New(width, height, isPrimeKind),
		vTree:            sessiontree.New(isPrimeKind),
		mode:             normalMode,
		modeStates: map[mode]*modeState{
			normalMode: {prompt: normalModePrompt},
			renameMode: {prompt: renameModePrompt},
			newMode:    {prompt: newModePrompt},
		},
	}
}

func (tui *TUI) Start() {
	go tui.eventReciever()
}

func (tui *TUI) SetOutputCh(outputEventCh chan event.Event) {
	tui.eventOutputCh = outputEventCh
}

func (tui *TUI) GetInputCh() chan event.Event {
	return tui.eventInputCh
}

func (tui *TUI) Render() {
	logger.Info("render")

	ms := tui.modeStates[tui.mode]
	frame := tui.printer.GenerateFrame(tui.vTree, ms.getPrompt()+string(ms.input))
	fmt.Print(frame) //nolint:forbidigo

	logger.Info("rendered")
}

func (tui *TUI) requestSessionPreview(sessionID string) {
	if tui.kind == PrimeKind {
		return
	}

	session := tui.sessionIDToSession[sessionID]
	panesWithoutSnapshot := session.GetPanesWithoutSnapshot()

	for _, pane := range panesWithoutSnapshot {
		tui.sendEvent(event.Event{
			Type: event.TypeCapturePane,
			Data: event.CapturePane{
				PaneID: pane.ID,
			},
		})
	}
}

func (tui *TUI) eventReciever() {
	for {
		inputEvent := <-tui.eventInputCh
		logger.Info("received event", slog.String("type", string(inputEvent.Type)))

		switch inputEvent.Type {
		case event.TypeKeyPressed:
			keyEvent, ok := inputEvent.Data.(event.KeyPressed)
			assert.Assert(ok, "Event data is not a EventKeyPressed")
			tui.handleKeyEvent(keyEvent)
		case event.TypeListedTree:
			sessions, ok := inputEvent.Data.(event.ListedTree)
			assert.Assert(ok, "Event data is not a TmuxCommandListTree")
			tui.handleListedTree(sessions.Sessions)
		case event.TypeListedFolders:
			sessions, ok := inputEvent.Data.([]*session.Session)
			assert.Assert(ok, "Event data is not a EventListedFolders")
			tui.handleListedFolders(sessions)
		case event.TypeCapturedPane:
			eventPane, ok := inputEvent.Data.(event.CapturedPane)
			assert.Assert(ok, "Event data is not a EventCapturedPane")
			tui.handleCapturedPane(eventPane.PaneID, eventPane.Snapshot)
		case event.TypeFetchedCurrentWindow:
			eventWindow, ok := inputEvent.Data.(event.FetchedCurrentWindow)
			assert.Assert(ok, "Event data is not a EventFetchedCurrentWindow")
			tui.handleFetchedCurrentWindow(eventWindow.SessionID, eventWindow.WindowID)
		default:
			assert.Fatal("Unknown event type")
		}
	}
}

func (tui *TUI) sendEvent(e event.Event) {
	tui.eventOutputCh <- e
}

//nolint:cyclop,gocognit,funlen
func (tui *TUI) handleCommand(input string, isDelete bool) {
	selectedSession := tui.vTree.GetSelectedSession()
	selectedWindow := tui.vTree.GetSelectedWindow()

	if os.Getenv("TMUX") == "" {
		os.Exit(1)
	}

	if tui.kind == PrimeKind {
		if selectedSession == nil {
			return
		}

		if strings.HasPrefix(selectedSession.ID, "notexisted_") {
			tmux.CreateTmuxSession(selectedSession.Name, selectedSession.Directory)
			tmux.SwitchClient(selectedSession.Name)

			os.Exit(0)
		}

		tmux.SwitchClient(selectedSession.ID)

		return
	}

	switch tui.mode {
	case normalMode:
		if !isDelete {
			sessionName := input

			if selectedSession != nil {
				entityID := selectedSession.ID
				if selectedWindow != nil {
					entityID = selectedWindow.ID
				}

				tmux.SwitchClient(entityID)
				os.Exit(0)
			}

			tmux.CreateTmuxSession(sessionName, tui.directory)
			tmux.SwitchClient(sessionName)
			os.Exit(0)
		}

		if selectedSession != nil {
			if selectedWindow != nil && len(selectedSession.Windows) > 1 {
				logger.Info("kill window", slog.String("sessionID", selectedSession.ID), slog.String("windowID", selectedWindow.ID), slog.String("windowName", selectedWindow.Name), slog.String("sesionName", selectedSession.Name))
				tmux.KillTmuxWindow(selectedWindow.ID)

				newWindows := make([]session.Window, 0)
				session := tui.sessionIDToSession[selectedSession.ID]

				for _, window := range session.Windows {
					if window.ID != selectedWindow.ID {
						newWindows = append(newWindows, window)
					}
				}

				session.Windows = newWindows

				return
			}

			logger.Info("kill session", slog.String("sessionID", selectedSession.ID), slog.String("sessionName", selectedSession.Name))
			tmux.KillTmuxSession(selectedSession.ID)

			newSessions := make([]*session.Session, 0)

			for _, session := range tui.sessions {
				if session.ID != selectedSession.ID {
					newSessions = append(newSessions, session)
				}
			}

			tui.sessions = newSessions
		}
	case renameMode:
		if input != "" {
			if selectedWindow != nil {
				tmux.RenameTmuxWindow(selectedWindow.ID, input)
				selectedWindow.Name = input
				session := tui.sessionIDToSession[selectedSession.ID]

				for i := range session.Windows {
					if session.Windows[i].ID == selectedWindow.ID {
						session.Windows[i].Name = input

						break
					}
				}

				return
			}

			tmux.RenameTmuxSession(selectedSession.ID, input)
			session := tui.sessionIDToSession[selectedSession.ID]
			session.Name = input
			selectedSession.Name = input
		}
	case newMode:
		if input != "" {
			tmux.CreateTmuxSession(input, tui.directory)
			tmux.SwitchClient(input)
			os.Exit(0)
		}
	}
}
