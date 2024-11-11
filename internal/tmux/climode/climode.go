package climode

import (
	"log/slog"
	"os/exec"
	"strings"

	"github.com/verte-zerg/gession/internal/event"
	"github.com/verte-zerg/gession/internal/tmux"
	"github.com/verte-zerg/gession/pkg/assert"
	"github.com/verte-zerg/gession/pkg/logging"
)

var (
	logger = logging.GetInstance().WithGroup("tmux_cli_mode")
)

type CLIMode struct {
	inputEventCh  chan event.Event
	outputEventCh chan event.Event

	commandsCh chan tmux.Command
	resultsCh  chan tmux.Command
}

func New() *CLIMode {
	return &CLIMode{
		commandsCh:    make(chan tmux.Command, event.MaxQueue),
		resultsCh:     make(chan tmux.Command, event.MaxQueue),
		inputEventCh:  make(chan event.Event, event.MaxQueue),
		outputEventCh: make(chan event.Event, event.MaxQueue),
	}
}

func (t *CLIMode) Start() {
	logger.Info("starting tmux command handler")

	go t.eventReceiver()
	go t.eventSender()
	go t.handler()
}

func (t *CLIMode) GetInputCh() chan event.Event {
	return t.inputEventCh
}

func (t *CLIMode) SetOutputCh(outputCh chan event.Event) {
	t.outputEventCh = outputCh
}

func (t *CLIMode) eventReceiver() {
	for {
		e := <-t.inputEventCh

		command := tmux.ConvertEventToCommand(e)
		t.commandsCh <- command
	}
}

func (t *CLIMode) eventSender() {
	for {
		command := <-t.resultsCh
		t.outputEventCh <- tmux.ConvertCommandToEvent(command)
	}
}

func (t *CLIMode) handler() {
	for {
		command := <-t.commandsCh

		commandLine := command.GetCommand(false)
		logger.Info("executing tmux command", slog.String("args", commandLine))

		args := strings.Split(commandLine, " ")

		cmd := exec.Command("tmux", args...)

		stdout, err := cmd.Output()
		assert.Assert(err == nil, "Failed to execute tmux command output: %s", commandLine)

		command.SetResult(string(stdout))

		t.resultsCh <- command
	}
}
