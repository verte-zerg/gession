package main

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

type Session struct {
	Name       string
	IsAttached bool
}

func getTmuxSessionList() ([]Session, error) {
	cmd := exec.Command("tmux", "ls", "-F", "#{?session_last_attached,#{session_last_attached},0000000000} #{?session_attached,*, } #{session_name}")
	stdout, err := cmd.Output()

	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(stdout), "\n")
	lines = lines[:len(lines)-1]

	slices.Sort(lines)
	slices.Reverse(lines)

	sessions := make([]Session, 0)
	sessionsAttached := make([]Session, 0)

	for _, line := range lines {
		if len(line) < 14 {
			continue
		}

		name := line[13:]
		if line[11] == '*' {
			sessionsAttached = append(sessionsAttached, Session{Name: name, IsAttached: true})
		} else {
			sessions = append(sessions, Session{Name: name, IsAttached: false})
		}
	}

	sessions = append(sessions, sessionsAttached...)
	return sessions, nil
}

func createTmuxSession(name string, directory string) {
	tmux := exec.Command("tmux", "new-session", "-d", "-s", name, "-c", directory)
	err := tmux.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func switchTmuxSession(name string) {
	tmux := exec.Command("tmux", "switch-client", "-t", name)
	err := tmux.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func killTmuxSession(name string) {
	tmux := exec.Command("tmux", "kill-session", "-t", name)
	err := tmux.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
