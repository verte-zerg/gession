package key

import (
	"github.com/verte-zerg/gession/pkg/logging"
	"unicode"
)

type Special string

var (
	logger = logging.GetInstance().WithGroup("key")
)

const (
	Usual     Special = "Usual"
	Up        Special = "Up" //nolint:varnamelen
	Down      Special = "Down"
	Left      Special = "Left"
	Right     Special = "Right"
	Esc       Special = "Esc"
	Backspace Special = "Backspace"
	Enter     Special = "Enter"
	Tab       Special = "Tab"
	ShiftTab  Special = "ShiftTab"
	ETX       Special = "ETX"
	EOT       Special = "EOT"
	Ignore    Special = "Ignore"
	CtrlR     Special = "CtrlR"
	CtrlT     Special = "CtrlT"
	CtrlE     Special = "CtrlE"

	escChar       byte = 27
	backspaceChar byte = 127
	enterChar     byte = 13
	tabChar       byte = 9
	etxChar       byte = 3
	eotChar       byte = 4
	ctrlRChar     byte = 18
	ctrlTChar     byte = 20
	ctrlEChar     byte = 5

	controlSeqLen = 3
)

type Key struct {
	Key        rune
	SpecialKey Special
}

//nolint:cyclop
func New(buf []byte, size int) Key {
	char := buf[0]

	logger.Debug("key", "char", char, "size", size, "buf", buf)

	if size == 1 {
		switch char {
		case escChar:
			return Key{SpecialKey: Esc}
		case backspaceChar:
			return Key{SpecialKey: Backspace}
		case enterChar:
			return Key{SpecialKey: Enter}
		case tabChar:
			return Key{SpecialKey: Tab}
		case etxChar:
			return Key{SpecialKey: ETX}
		case eotChar:
			return Key{SpecialKey: EOT}
		case ctrlRChar:
			return Key{SpecialKey: CtrlR}
		case ctrlTChar:
			return Key{SpecialKey: CtrlT}
		case ctrlEChar:
			return Key{SpecialKey: CtrlE}
		default:
			value := []rune(string(char))[0]
			if unicode.IsPrint(value) {
				return Key{Key: []rune(string(char))[0], SpecialKey: Usual}
			}

			return Key{SpecialKey: Ignore}
		}
	} else if size == controlSeqLen {
		if buf[0] == escChar && buf[1] == '[' {
			switch buf[2] {
			case 'A':
				return Key{SpecialKey: Up}
			case 'B':
				return Key{SpecialKey: Down}
			case 'C':
				return Key{SpecialKey: Right}
			case 'D':
				return Key{SpecialKey: Left}
			case 'Z':
				return Key{SpecialKey: ShiftTab}
			}

			return Key{SpecialKey: Ignore}
		}
	}

	value := []rune(string(buf))[0]
	if unicode.IsPrint(value) {
		return Key{Key: value, SpecialKey: Usual}
	}

	return Key{SpecialKey: Ignore}
}
