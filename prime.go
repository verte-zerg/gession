package main

import (
	"fmt"
	"os"
	"path"
)

func getPrimeagenSessionList(dirs []string) ([]Session, error) {
	var sessions []Session
	for _, dir := range dirs {
		files, err := os.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("could not read directory %s: %w", dir, err)
		}

		for _, file := range files {
			if file.IsDir() {
				fullPath := path.Join(dir, file.Name())
				sessions = append(sessions, Session{
					Name:       file.Name(),
					IsAttached: false,
					Directory:  fullPath,
				})
			}
		}
	}

	return sessions, nil
}
