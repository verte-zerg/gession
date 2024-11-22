package commandmode

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/verte-zerg/gession/internal/event"
	"github.com/verte-zerg/gession/internal/tmux"
	"github.com/verte-zerg/gession/pkg/logging"
)

var (
	logger = logging.GetInstance().WithGroup("tmux_command_mode")
)

type CommandMode struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	scanner *bufio.Scanner

	inputEventCh  chan event.Event
	outputEventCh chan event.Event

	commandsCh       chan tmux.Command
	repeatCommandsCh chan tmux.Command
	resultsCh        chan tmux.Command
}

func New() *CommandMode {
	return &CommandMode{
		commandsCh:       make(chan tmux.Command, event.MaxQueue),
		repeatCommandsCh: make(chan tmux.Command, event.MaxQueue),
		resultsCh:        make(chan tmux.Command, event.MaxQueue),
		inputEventCh:     make(chan event.Event, event.MaxQueue),
	}
}

func (t *CommandMode) Start() error {
	logger.Info("starting tmux API")

	cmd := exec.Command("tmux", "-C", "attach")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("could not get stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("could not get stdout pipe: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("could not start tmux: %w", err)
	}

	scanner := bufio.NewScanner(stdout)

	t.cmd = cmd
	t.stdin = stdin
	t.stdout = stdout
	t.scanner = scanner

	go t.commandReciever()
	go t.commandHandler()
	go t.eventReciever()
	go t.eventSender()

	return nil
}

func (t *CommandMode) GetInputCh() chan event.Event {
	return t.inputEventCh
}

func (t *CommandMode) SetOutputCh(outputCh chan event.Event) {
	logger.Info("set output event channel for tmux")

	t.outputEventCh = outputCh
}

func (t *CommandMode) eventReciever() {
	for {
		e := <-t.inputEventCh

		command := tmux.ConvertEventToCommand(e)
		t.commandsCh <- command
		t.repeatCommandsCh <- command
	}
}

func (t *CommandMode) eventSender() {
	for {
		command := <-t.resultsCh
		t.outputEventCh <- tmux.ConvertCommandToEvent(command)
	}
}

func (t *CommandMode) commandReciever() {
	for {
		command := <-t.commandsCh

		commandLine := command.GetCommand(true)
		logger.Info("start tmux command", slog.String("command", commandLine))
		_, err := t.stdin.Write([]byte(commandLine + "\n"))

		if err != nil {
			panic(err)
		}
	}
}

func (t *CommandMode) commandHandler() {
	isCommandStarted := false
	expectedEndLine := ""

	commandResult := strings.Builder{}
	isFirstCommand := true

	for t.scanner.Scan() {
		line := t.scanner.Text()

		if isCommandStarted {
			if line == expectedEndLine {
				if isFirstCommand {
					isFirstCommand = false
					isCommandStarted = false

					continue
				}

				currentCommand := <-t.repeatCommandsCh
				commandLine := currentCommand.GetCommand(true)
				logger.Info("finish tmux command", slog.String("command", commandLine))

				currentCommand.SetResult(commandResult.String())
				t.resultsCh <- currentCommand

				isCommandStarted = false

				commandResult.Reset()

				continue
			}

			commandResult.WriteString(line + "\n")
		}

		if !isCommandStarted && strings.HasPrefix(line, "%begin") {
			expectedEndLine = strings.ReplaceAll(line, "%begin", "%end")
			isCommandStarted = true

			continue
		}
	}
}
