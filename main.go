package main

import (
	"flag"
	"fmt"
	"golang.org/x/term"
	"os"
)

const (
	DEFAULT_PROMPT = "input > "
)

type CmdArgs struct {
	Prompt    string
	Directory string
}

func parseArgs() (*CmdArgs, error) {
	prompt := flag.String("p", DEFAULT_PROMPT, "prompt to display")
	directory := flag.String("d", "", "directory to start in (default: \"$HOME\")")

	flag.Parse()

	if *directory == "" {
		homedir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		*directory = homedir
	}

	return &CmdArgs{
		Prompt:    *prompt,
		Directory: *directory,
	}, nil
}

func main() {
	cmdArgs, err := parseArgs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Print("\033[H\033[2J\033[999H")

	fd := int(os.Stdin.Fd())
	_, height, err := term.GetSize(fd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	sessions, err := getTmuxSessionList()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tui := NewTUI(height, sessions, cmdArgs.Prompt, cmdArgs.Directory)
	tui.Render("")

	keyCh := make(chan KeyEvent)
	go captureKeys(keyCh)

	for {
		key := <-keyCh
		tui.Iterate(key)
	}
}
