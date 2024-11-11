package sessiontree

import (
	"strings"

	"github.com/verte-zerg/gession/internal/session"
	"github.com/verte-zerg/gession/pkg/assert"
	"github.com/verte-zerg/gession/pkg/fuzzy"
)

const (
	BOLD  = "\033[1m"
	GREEN = "\033[32m"
	RESET = "\033[0m"

	PARTS = 3
)

type VisualizeTree struct {
	showEmptyEntities bool

	tree            []FilteredSession
	selectedSession *FilteredSession
	selectedWindow  *FilteredWindow

	visibleRows            int
	selectedIdx            int
	firstSelectedWindowIdx int

	sessionsCount int
}

func New(showEmptyEntities bool) *VisualizeTree {
	return &VisualizeTree{
		showEmptyEntities: showEmptyEntities,
		tree:              nil,
	}
}

func (vt VisualizeTree) GetSelectedSession() *FilteredSession {
	return vt.selectedSession
}

func (vt VisualizeTree) GetSelectedWindow() *FilteredWindow {
	return vt.selectedWindow
}

func (vt VisualizeTree) GetVisibleRows() int {
	return vt.visibleRows
}

func (vt VisualizeTree) GetSelectedIdx() int {
	return vt.selectedIdx
}

func (vt VisualizeTree) GetFirstSelectedWindowIdx() int {
	return vt.firstSelectedWindowIdx
}

func (vt VisualizeTree) GetSessionsCount() int {
	return vt.sessionsCount
}

func (vt VisualizeTree) GetSessions() []FilteredSession {
	return vt.tree
}

type FilteredSession struct {
	*session.Session

	IsUnwrapped      bool
	FilteredChildren []*FilteredWindow
	query            string
}

func (s FilteredSession) GetString(bold bool) string {
	return getRepresentation(s.Name, s.query, bold)
}

type FilteredWindow struct {
	*session.Window

	FilteredChildren []*FilteredPane
	query            string
}

func (w FilteredWindow) GetString(bold bool) string {
	return getRepresentation(w.Name, w.query, bold)
}

type FilteredPane struct {
	*session.Pane

	query string
}

func (p FilteredPane) GetString(bold bool) string {
	return getRepresentation(p.CurrentCommand, p.query, bold)
}

func getRepresentation(name, query string, bold bool) string {
	boldMarker := ""
	if bold {
		boldMarker = BOLD
	}

	builder := strings.Builder{}
	representation, _ := fuzzy.SearchColorized(name, query)

	for _, c := range representation {
		if c.Highlighted {
			builder.WriteString(boldMarker + GREEN + c.Text + RESET)
		} else {
			builder.WriteString(boldMarker + c.Text + RESET)
		}
	}

	return builder.String()
}

//nolint:cyclop
func (vt *VisualizeTree) SearchEntities(
	query string,
	sessions []*session.Session,
	selectedIdx int,
	unwrappedSession map[string]interface{},
) []FilteredSession {
	vt.sessionsCount = len(sessions)
	query = strings.TrimSpace(query)
	vt.selectedIdx = selectedIdx

	queryParts := strings.Split(query, " ")
	if len(queryParts) > PARTS {
		queryParts = queryParts[:PARTS]
	}

	for len(queryParts) < PARTS {
		queryParts = append(queryParts, "")
	}

	tree := make([]FilteredSession, 0)
	visibleRows := 0

	for _, session := range sessions {
		if !fuzzy.Search(session.Name, queryParts[0]) {
			continue
		}

		isUnwrapped := unwrappedSession[session.ID] != nil

		filteredSession := FilteredSession{
			Session:     session,
			IsUnwrapped: isUnwrapped,
			query:       queryParts[0],
		}

		for _, window := range session.Windows {
			if !fuzzy.Search(window.Name, queryParts[1]) {
				continue
			}

			filteredWindow := FilteredWindow{
				Window: &window,
				query:  queryParts[1],
			}

			for _, pane := range window.Panes {
				if !fuzzy.Search(pane.CurrentCommand, queryParts[2]) {
					continue
				}

				filteredPane := FilteredPane{
					Pane:  &pane,
					query: queryParts[2],
				}

				filteredWindow.FilteredChildren = append(filteredWindow.FilteredChildren, &filteredPane)
			}

			if len(filteredWindow.FilteredChildren) != 0 || vt.showEmptyEntities {
				filteredSession.FilteredChildren = append(filteredSession.FilteredChildren, &filteredWindow)

				if isUnwrapped {
					visibleRows++
				}
			}
		}

		if len(filteredSession.FilteredChildren) != 0 || vt.showEmptyEntities {
			tree = append(tree, filteredSession)
			visibleRows++
		}
	}

	vt.visibleRows = visibleRows
	vt.tree = tree
	vt.markSelectedEntities()

	return tree
}

func (vt *VisualizeTree) markSelectedEntities() {
	vt.selectedSession = nil
	vt.selectedWindow = nil
	vt.firstSelectedWindowIdx = 0

	if len(vt.tree) == 0 {
		return
	}

	if vt.selectedIdx < 0 {
		vt.selectedIdx = 0
	}

	if vt.selectedIdx >= vt.visibleRows {
		vt.selectedIdx = vt.visibleRows - 1
	}

	orderIdx := 0

	for sessionIdx, session := range vt.tree {
		if session.IsUnwrapped {
			if len(session.FilteredChildren) != 0 {
				vt.firstSelectedWindowIdx = orderIdx
			}

			for windowIdx := range session.FilteredChildren {
				if orderIdx == vt.selectedIdx {
					vt.selectedSession = &vt.tree[sessionIdx]
					vt.selectedWindow = vt.tree[sessionIdx].FilteredChildren[windowIdx]

					return
				}

				orderIdx++
			}
		}

		if orderIdx == vt.selectedIdx {
			vt.selectedSession = &vt.tree[sessionIdx]

			return
		}

		orderIdx++
	}

	assert.Fatal("Selected session not found")
}
