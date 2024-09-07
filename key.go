package main

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

type SpecialKey int

const (
	// Special keys
	USUAL SpecialKey = iota
	UP
	DOWN
	ESC
	BACKSPACE
	ENTER
	ETX
	EOT

	// Usual keys
	ESC_CHAR       byte = 27
	BACKSPACE_CHAR      = 127
	ENTER_CHAR          = 13
	ETX_CHAR            = 3
	EOT_CHAR            = 4
)

type KeyEvent struct {
	Key        rune
	SpecialKey SpecialKey
	char       byte
	buf        []byte
	size       int
}

func (k *KeyEvent) DetectKey() {
	if k.size == 1 {
		switch k.char {
		case ESC_CHAR:
			k.SpecialKey = ESC
		case BACKSPACE_CHAR:
			k.SpecialKey = BACKSPACE
		case ENTER_CHAR:
			k.SpecialKey = ENTER
		case ETX_CHAR:
			k.SpecialKey = ETX
		case EOT_CHAR:
			k.SpecialKey = EOT
		default:
			k.Key = []rune(string(k.buf))[0]
		}
		return
	} else if k.size == 3 {
		if k.buf[0] == 27 && k.buf[1] == 91 {
			switch k.buf[2] {
			case 65:
				k.SpecialKey = UP
			case 66:
				k.SpecialKey = DOWN
			}
			if k.buf[2] == 65 || k.buf[2] == 66 {
				return
			}
		}
	}
	k.Key = []rune(string(k.buf))[0]
}

func NewKey(buf []byte, size int) KeyEvent {
	b := make([]byte, 4)
	copy(b, buf)
	keyEvent := KeyEvent{
		char: b[0],
		buf:  b,
		size: size,
	}
	keyEvent.DetectKey()
	return keyEvent
}

func captureKeys(ch chan KeyEvent) {
	fd := int(os.Stdin.Fd())
	t, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer term.Restore(fd, t)

	b := make([]byte, 4)
	for {
		for i := 0; i < len(b); i++ {
			b[i] = 0
		}
		n, err := os.Stdin.Read(b)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		ch <- NewKey(b, n)
	}
}
