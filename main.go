package main

import (
	"flag"
	"fmt"
	"os"

	"golang.org/x/term"
)

const (
	DEFAULT_PROMPT = "input > "
)

type Mode int

const (
	ModeNormal Mode = iota
	ModePrime
)

type CmdArgs struct {
	Prompt    string
	Directory string
	PrimeDirs []string
	Mode      Mode
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func parseArgs() (*CmdArgs, error) {
	prompt := flag.String("i", DEFAULT_PROMPT, "prompt to display")
	directory := flag.String("d", "", "directory to start in (default: \"$HOME\")")
	mode := ModeNormal
	isPrime := flag.Bool("prime", false, "start in primeagen mode. This will not show tmux sessions, but folders in the specified directories")
	primeDirs := arrayFlags{}
	flag.Var(&primeDirs, "pd", "directories to search for primeagen sessions. Can be specified multiple times")
	flag.Parse()

	fmt.Println(primeDirs)

	if *isPrime {
		mode = ModePrime
	}

	switch mode {
	case ModeNormal:
		if len(primeDirs) != 0 {
			return nil, fmt.Errorf("primeagen directories can only be specified in primeagen mode")
		}

		if *directory == "" {
			homedir, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("could not determine home directory: %v", err)
			}
			*directory = homedir
		}

		return &CmdArgs{
			Prompt:    *prompt,
			Directory: *directory,
			Mode:      mode,
			PrimeDirs: nil,
		}, nil

	case ModePrime:
		if len(primeDirs) == 0 {
			return nil, fmt.Errorf("primeagen mode requires at least one directory to search for sessions")
		}

		primeDirsList := make([]string, 0, len(primeDirs))
		for _, dir := range primeDirs {
			if _, err := os.Stat(dir); err != nil {
				return nil, fmt.Errorf("directory %s does not exist", dir)
			}
			primeDirsList = append(primeDirsList, dir)
		}

		return &CmdArgs{
			Prompt:    *prompt,
			Directory: "",
			Mode:      mode,
			PrimeDirs: primeDirsList,
		}, nil
	default:
		return nil, fmt.Errorf("unknown mode")
	}
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

	var sessions []Session
	switch cmdArgs.Mode {
	case ModeNormal:
		sessions, err = getTmuxSessionList(cmdArgs.Directory)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case ModePrime:
		sessions, err = getPrimeagenSessionList(cmdArgs.PrimeDirs)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if len(sessions) == 0 {
			fmt.Println("no folders found in specified directories")
			os.Exit(1)
		}
	}

	tui := NewTUI(height, sessions, cmdArgs.Mode, cmdArgs.Prompt, cmdArgs.Directory)
	tui.Render("")

	keyCh := make(chan KeyEvent)
	go captureKeys(keyCh)

	for {
		key := <-keyCh
		tui.Iterate(key)
	}
}
