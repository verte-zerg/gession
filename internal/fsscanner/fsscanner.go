package fsscanner

import (
	"os"
	"path"
	"slices"
	"sort"
	"strings"

	"github.com/verte-zerg/gession/internal/event"
	"github.com/verte-zerg/gession/internal/session"
	"github.com/verte-zerg/gession/pkg/assert"
	"github.com/verte-zerg/gession/pkg/logging"
)

var (
	logger = logging.GetInstance().WithGroup("fsscanner")
)

type FSScanner struct {
	inputEventCh  chan event.Event
	outputEventCh chan event.Event
}

func New() *FSScanner {
	return &FSScanner{
		inputEventCh:  make(chan event.Event, event.MaxQueue),
		outputEventCh: make(chan event.Event, event.MaxQueue),
	}
}

func (t *FSScanner) Start() {
	logger.Info("starting fsscanner handler")

	go t.handler()
}

func (t *FSScanner) GetInputCh() chan event.Event {
	return t.inputEventCh
}

func (t *FSScanner) SetOutputCh(outputCh chan event.Event) {
	t.outputEventCh = outputCh
}

func (t *FSScanner) handler() {
	for {
		e := <-t.inputEventCh

		assert.Assert(e.Type == event.TypeListFolders, "fsscanner supports only list folders event")

		folders, ok := e.Data.([]string)
		assert.Assert(ok, "data should be a list of strings")

		pathsMap := make(map[string]struct{})

		slices.Reverse(folders)

		for _, folder := range folders {
			for _, subfolder := range listFolder(folder) {
				pathsMap[subfolder] = struct{}{}
			}
		}

		sessions := make([]*session.Session, 0, len(pathsMap))
		for subfolder := range pathsMap {
			sessions = append(sessions, convertFolderToSession(subfolder))
		}

		sort.Slice(sessions, func(i, j int) bool {
			return sessions[i].Name > sessions[j].Name
		})

		t.outputEventCh <- event.Event{
			Type: event.TypeListedFolders,
			Data: sessions,
		}
	}
}

func listFolder(folder string) []string {
	entities, err := os.ReadDir(folder)
	assert.Assert(err == nil, "could not read directory")

	files := make([]string, 0)

	for _, entity := range entities {
		if entity.IsDir() {
			files = append(files, path.Join(folder, entity.Name()))
		}
	}

	return files
}

func convertFolderToSession(folderPath string) *session.Session {
	dirname := path.Base(folderPath)

	normalizedDirname := strings.ReplaceAll(dirname, ".", "_")

	return &session.Session{
		ID:        "notexisted_" + dirname,
		Name:      normalizedDirname,
		Directory: folderPath,
	}
}
