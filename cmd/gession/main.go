package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/verte-zerg/gession/internal/event"
	"github.com/verte-zerg/gession/internal/fsscanner"
	"github.com/verte-zerg/gession/internal/keyboard"
	"github.com/verte-zerg/gession/internal/tmux/climode"
	"github.com/verte-zerg/gession/internal/tmux/commandmode"
	"github.com/verte-zerg/gession/internal/tui"
	"github.com/verte-zerg/gession/pkg/assert"
	"github.com/verte-zerg/gession/pkg/logging"

	"golang.org/x/term"
)

var (
	logger = logging.GetInstance().WithGroup("main")
)

type CmdArgs struct {
	Directory string
	PrimeDirs []string
	Legacy    bool
	Prime     bool
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
	directory := flag.String("d", "", "directory to start a new session (default: \"$HOME\")")
	legacy := flag.Bool("legacy", false, "use tmux CLI instead of API to get session/buffer list")
	prime := flag.Bool("prime", false, "prime mode")
	primeDirs := arrayFlags{}
	flag.Var(&primeDirs, "pd", "directories to search for primeagen mode. Can be specified multiple times")

	flag.Parse()

	primeDirsList := make([]string, 0, len(primeDirs))

	if *prime {
		*directory = "/"

		assert.Assert(len(primeDirs) > 0, "no prime directories specified")

		for _, dir := range primeDirs {
			_, err := os.Stat(dir)
			assert.Assert(err == nil, "directory %s does not exist", dir)
			primeDirsList = append(primeDirsList, dir)
		}
	}

	if *directory == "" {
		homedir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("could not determine home directory: %w", err)
		}

		*directory = homedir
	}

	return &CmdArgs{
		Directory: *directory,
		PrimeDirs: primeDirsList,
		Legacy:    *legacy,
		Prime:     *prime,
	}, nil
}

func initEventSystem(tuiCP event.ConsumerProducer, tmuxCP event.ConsumerProducer, keyboardP event.Producer, fsscannerCP event.ConsumerProducer) *event.Router {
	eventSystem := event.New()
	eventSystem.RegisterConsumer([]event.Type{
		event.TypeKeyPressed,
		event.TypeCapturedPane,
		event.TypeListedTree,
		event.TypeListedFolders,
		event.TypeFetchedCurrentWindow,
	}, tuiCP)
	eventSystem.RegisterConsumer([]event.Type{
		event.TypeListTree,
		event.TypeCapturePane,
		event.TypeFetchCurrentWindow,
	}, tmuxCP)
	eventSystem.RegisterConsumer([]event.Type{
		event.TypeListFolders,
	}, fsscannerCP)
	eventSystem.RegisterProducer(tmuxCP)
	eventSystem.RegisterProducer(tuiCP)
	eventSystem.RegisterProducer(keyboardP)
	eventSystem.RegisterProducer(fsscannerCP)

	eventSystem.Start()

	return eventSystem
}

func emitInitialEvents(router *event.Router, prime bool, primeDirs []string) {
	router.EmitEvent(event.Event{
		Type: event.TypeListTree,
	})

	if prime {
		router.EmitEvent(event.Event{
			Type: event.TypeListFolders,
			Data: primeDirs,
		})
	} else {
		router.EmitEvent(event.Event{
			Type: event.TypeFetchCurrentWindow,
		})
	}
}

func initTUI(width, height int, tuiKind tui.Kind, directory string) *tui.TUI {
	tui := tui.NewTUI(width, height, tuiKind, directory)
	tui.Start()

	return tui
}

func initTmuxCommandMode() *commandmode.CommandMode {
	t := commandmode.New()
	err := t.Start()
	assert.Assert(err == nil, "could not start tmux in COMMAND mode")

	return t
}

func initTmuxCLIMode() *climode.CLIMode {
	t := climode.New()
	t.Start()

	return t
}

func initFSScanner() *fsscanner.FSScanner {
	f := fsscanner.New()
	f.Start()

	return f
}

func initKeyboard() *keyboard.Keyboard {
	keyboard := keyboard.NewKeyboard()
	keyboard.Start()

	return keyboard
}

func main() {
	logger.Info("starting gession")

	cmdArgs, err := parseArgs()
	assert.Assert(err == nil, "could not parse args: %v", err)

	// Clear the screen
	fmt.Print("\033[H\033[2J") //nolint:forbidigo

	fd := int(os.Stdin.Fd())
	width, height, err := term.GetSize(fd)
	assert.Assert(err == nil, "could not get terminal size: %v", err)

	kind := tui.NormalKind
	if cmdArgs.Prime {
		kind = tui.PrimeKind
	}

	tui := initTUI(width, height, kind, cmdArgs.Directory)
	scanner := initFSScanner()
	keyboard := initKeyboard()

	var tmuxInterface event.ConsumerProducer

	if cmdArgs.Legacy {
		tmuxInterface = initTmuxCLIMode()
	} else {
		tmuxInterface = initTmuxCommandMode()
	}

	router := initEventSystem(tui, tmuxInterface, keyboard, scanner)
	emitInitialEvents(router, cmdArgs.Prime, cmdArgs.PrimeDirs)

	logger.Info("waiting for events")
	select {}
}
