package session

import (
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/verte-zerg/gession/pkg/logging"
)

var (
	logger = logging.GetInstance().WithGroup("session")
)

type Pane struct {
	ID             string
	CurrentCommand string
	Index          int
	IsActive       bool
	Snapshot       *string
}

type Window struct {
	ID       string
	Name     string
	Index    int
	IsActive bool
	Panes    []Pane
}

type Session struct {
	ID               string
	Name             string
	IsAttached       bool
	LastTimeAttached time.Time
	Windows          []Window
	Directory        string
}

func (s Session) GetPanesWithoutSnapshot() []Pane {
	panes := make([]Pane, 0)

	for _, window := range s.Windows {
		for _, pane := range window.Panes {
			if pane.Snapshot == nil || *pane.Snapshot == "" {
				panes = append(panes, pane)
			}
		}
	}

	return panes
}

func (s *Session) SetSnapshot(paneID string, snapshot string) {
	for i, window := range s.Windows {
		for j, pane := range window.Panes {
			if pane.ID == paneID {
				*s.Windows[i].Panes[j].Snapshot = snapshot

				return
			}
		}
	}
}

type tmuxPaneResponse struct {
	sessionName        string
	windowName         string
	paneCurrentCommand string
	windowIndex        int
	paneIndex          int
	sessionAttached    bool
	windowActive       bool
	paneActive         bool
	lastAttached       time.Time
	sessionID          string
	windowID           string
	paneID             string
}

func parseResponse(response string) (*tmuxPaneResponse, error) {
	parts := strings.Split(response, "|")
	sessionName := parts[0]
	windowName := parts[1]
	paneCurrentCommand := parts[2]

	indices := strings.Split(parts[3], ".")
	windowIndex, err := strconv.Atoi(indices[0])

	if err != nil {
		return nil, fmt.Errorf("could not convert window index to int: %w", err)
	}

	paneIndex, err := strconv.Atoi(indices[1])
	if err != nil {
		return nil, fmt.Errorf("could not convert pane index to int: %w", err)
	}

	attachedParts := strings.Split(parts[4], ".")
	sessionAttached := attachedParts[0] != "0"
	windowActive := attachedParts[1] == "1"
	paneActive := attachedParts[2] == "1"

	lastAttached := time.Time{}

	if len(parts[5]) > 0 {
		ts, err := strconv.ParseInt(parts[5], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("could not convert last attached timestamp to int: %w", err)
		}

		lastAttached = time.Unix(ts, 0)
	}

	ids := parts[6]
	idParts := strings.Split(ids, ".")
	sessionID := idParts[0]
	windowID := idParts[1]
	paneID := idParts[2]

	logger.Info("Parsed response", slog.String("sessionName", sessionName), slog.String("windowName", windowName), slog.String("paneID", paneID))

	return &tmuxPaneResponse{
		sessionName:        sessionName,
		windowName:         windowName,
		paneCurrentCommand: paneCurrentCommand,
		windowIndex:        windowIndex,
		paneIndex:          paneIndex,
		sessionAttached:    sessionAttached,
		windowActive:       windowActive,
		paneActive:         paneActive,
		lastAttached:       lastAttached,
		paneID:             paneID,
		sessionID:          sessionID,
		windowID:           windowID,
	}, nil
}

func ParseSessions(response string) ([]*Session, error) {
	panes := strings.Split(response, "\n")

	rawSessions := make(map[string]map[int]map[int]*tmuxPaneResponse)

	for _, line := range panes {
		response, err := parseResponse(line)
		if err != nil {
			return nil, fmt.Errorf("could not parse response: %w", err)
		}

		if _, ok := rawSessions[response.sessionName]; !ok {
			rawSessions[response.sessionName] = make(map[int]map[int]*tmuxPaneResponse)
		}

		if _, ok := rawSessions[response.sessionName][response.windowIndex]; !ok {
			rawSessions[response.sessionName][response.windowIndex] = make(map[int]*tmuxPaneResponse)
		}

		rawSessions[response.sessionName][response.windowIndex][response.paneIndex] = response
	}

	sessionList := make([]*Session, 0, len(rawSessions))

	for sessionName, windows := range rawSessions {
		session := &Session{
			Name: sessionName,
		}

		for windowIndex, panes := range windows {
			window := Window{
				Index: windowIndex,
			}

			for paneIndex, response := range panes {
				emptySnapshot := ""
				pane := Pane{
					Index:          paneIndex,
					IsActive:       response.paneActive,
					CurrentCommand: response.paneCurrentCommand,
					ID:             response.paneID,
					Snapshot:       &emptySnapshot,
				}

				session.ID = response.sessionID
				session.IsAttached = response.sessionAttached
				session.LastTimeAttached = response.lastAttached
				window.ID = response.windowID
				window.Name = response.windowName
				window.Index = response.windowIndex
				window.IsActive = response.windowActive

				window.Panes = append(window.Panes, pane)
			}

			sort.Slice(window.Panes, func(i, j int) bool {
				return window.Panes[i].Index < window.Panes[j].Index
			})

			session.Windows = append(session.Windows, window)
		}

		sort.Slice(session.Windows, func(i, j int) bool {
			return session.Windows[i].Index < session.Windows[j].Index
		})

		sessionList = append(sessionList, session)
	}

	sort.Slice(sessionList, func(i, j int) bool {
		return sessionList[i].LastTimeAttached.After(sessionList[j].LastTimeAttached)
	})

	if len(sessionList) > 1 {
		sessionList = append(sessionList[1:], sessionList[0])
	}

	return sessionList, nil
}
