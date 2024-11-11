package keyboard

import (
	"log/slog"
	"os"

	"github.com/verte-zerg/gession/internal/event"
	"github.com/verte-zerg/gession/internal/key"
	"github.com/verte-zerg/gession/pkg/assert"
	"github.com/verte-zerg/gession/pkg/logging"

	"golang.org/x/term"
)

var (
	logger = logging.GetInstance().WithGroup("keyboard")
)

const (
	MaxKeyLength = 4
)

type Keyboard struct {
	outputEventCh chan event.Event
}

func NewKeyboard() *Keyboard {
	return &Keyboard{}
}

func (k *Keyboard) SetOutputCh(outputEventCh chan event.Event) {
	k.outputEventCh = outputEventCh
}

func (k *Keyboard) captureKeys() {
	fd := int(os.Stdin.Fd())
	t, err := term.MakeRaw(fd)
	assert.Assert(err == nil, "could not make raw terminal")

	defer func() {
		err := term.Restore(fd, t)
		assert.Assert(err == nil, "could not restore terminal")
	}()

	b := make([]byte, MaxKeyLength)

	for {
		for i := 0; i < len(b); i++ {
			b[i] = 0
		}

		n, err := os.Stdin.Read(b)

		assert.Assert(err == nil, "could not read from stdin")

		logger.Info("ket was pressed", slog.Any("key", b[:n]), slog.Int("length", n))
		key := key.New(b, n)
		k.outputEventCh <- event.Event{
			Type: event.TypeKeyPressed,
			Data: event.KeyPressed{
				SpecialKey: key.SpecialKey,
				Key:        key.Key,
			},
		}
	}
}

func (k *Keyboard) Start() {
	go k.captureKeys()
}
