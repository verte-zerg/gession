package tui

import (
	"github.com/verte-zerg/gession/internal/session"
	"log/slog"
)

func (tui *TUI) filterSessions() {
	normalInput := tui.modeStates[normalMode].input
	inputString := string(normalInput)

	logger.Info("searching entities", slog.String("input", inputString))
	tui.vTree.SearchEntities(inputString, tui.sessions, tui.selectedIdx, tui.unwrappedSession)

	tui.selectedIdx = tui.vTree.GetSelectedIdx()
	selectedSession := tui.vTree.GetSelectedSession()

	if selectedSession != nil {
		tui.requestSessionPreview(selectedSession.ID)
	}
}

func (tui *TUI) mergeSessionsAndPrimeSessions(primeSessions []*session.Session, normalSessions []*session.Session) {
	logger.Info("merging sessions and prime sessions")

	tui.sessions = primeSessions

	tui.sessionIDToSession = make(map[string]*session.Session)

	sessionNameToSession := make(map[string]*session.Session)
	for _, session := range normalSessions {
		sessionNameToSession[session.Name] = session
	}

	for _, primeSession := range primeSessions {
		if normalSession, ok := sessionNameToSession[primeSession.Name]; ok {
			logger.Info("merging session", slog.String("primeSessionID", primeSession.ID), slog.String("normalSessionID", normalSession.ID))
			primeSession.ID = normalSession.ID
		}

		tui.sessionIDToSession[primeSession.ID] = primeSession
	}

	logger.Info("searching entities", slog.String("input", ""))
	tui.vTree.SearchEntities("", tui.sessions, tui.selectedIdx, tui.unwrappedSession)

	tui.Render()
}

func (tui *TUI) handleListedFolders(sessions []*session.Session) {
	if tui.kind != PrimeKind {
		return
	}

	if tui.tmpSessions == nil {
		tui.tmpSessions = sessions
		logger.Info("tmp sessions", slog.Int("count", len(sessions)), slog.String("event_type", "listed_folders"))

		return
	}

	tui.mergeSessionsAndPrimeSessions(sessions, tui.tmpSessions)
}

func (tui *TUI) handleListedTree(sessions []*session.Session) {
	if tui.kind == PrimeKind {
		if tui.tmpSessions == nil {
			logger.Info("tmp sessions", slog.Int("count", len(sessions)), slog.String("event_type", "listed_tree"))
			tui.tmpSessions = sessions

			return
		}

		tui.mergeSessionsAndPrimeSessions(tui.tmpSessions, sessions)

		return
	}

	tui.sessions = sessions

	tui.paneIDToSession = make(map[string]*session.Session)
	tui.sessionIDToSession = make(map[string]*session.Session)

	for _, session := range sessions {
		logger.Info("session", slog.String("id", session.ID), slog.Bool("attached", session.IsAttached))

		tui.sessionIDToSession[session.ID] = session

		for j := range session.Windows {
			window := &session.Windows[j]

			for k := range window.Panes {
				pane := &window.Panes[k]
				tui.paneIDToSession[pane.ID] = session

				logger.Info(
					"list tree",
					slog.String("sessionID", session.ID),
					slog.String("sessionName", session.Name),
					slog.Bool("sessionAttached", session.IsAttached),
					slog.String("windowID", window.ID),
					slog.String("windowName", window.Name),
					slog.Bool("windowActive", window.IsActive),
					slog.String("paneID", pane.ID),
					slog.String("paneCommand", pane.CurrentCommand),
					slog.Bool("paneActive", pane.IsActive),
				)
			}
		}
	}

	logger.Info("searching entities", slog.String("input", ""))
	tui.vTree.SearchEntities("", tui.sessions, tui.selectedIdx, tui.unwrappedSession)
	tui.requestSessionPreview(sessions[0].ID)

	tui.Render()
}

func (tui *TUI) handleCapturedPane(paneID, snapshot string) {
	if session, ok := tui.paneIDToSession[paneID]; ok {
		session.SetSnapshot(paneID, snapshot)

		if len(session.GetPanesWithoutSnapshot()) == 0 {
			tui.Render()
		}
	}
}

func (tui *TUI) handleFetchedCurrentWindow(sessionID, windowID string) {
	logger.Info("current window", slog.String("sessionID", sessionID), slog.String("windowID", windowID))
	s := tui.sessionIDToSession[sessionID]
	windows := make([]session.Window, 0)

	for _, window := range s.Windows {
		if window.ID != windowID {
			windows = append(windows, window)
		}
	}

	s.Windows = windows

	logger.Info("searching entities", slog.String("input", ""))
	tui.vTree.SearchEntities("", tui.sessions, tui.selectedIdx, tui.unwrappedSession)
	tui.Render()
}
