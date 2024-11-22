package event

import (
	"github.com/verte-zerg/gession/internal/key"
	"github.com/verte-zerg/gession/internal/session"
)

const (
	TypeCapturePane   Type = Type("CapturePane")
	TypeCapturedPane  Type = Type("CapturedPane")
	TypeListTree      Type = Type("ListTree")
	TypeListedTree    Type = Type("ListedTree")
	TypeListFolders   Type = Type("ListFolders")
	TypeListedFolders Type = Type("ListedFolders")
	TypeKeyPressed    Type = Type("KeyPressed")
)

type Event struct {
	Type Type
	Data interface{}
}

type CapturePane struct {
	PaneID string
}

type CapturedPane struct {
	PaneID   string
	Snapshot string
}

type ListedTree struct {
	Sessions []*session.Session
}

type ListFolders struct {
	Folders []string
}

type ListedFolders struct {
	Sessions []*session.Session
}

type KeyPressed struct {
	SpecialKey key.Special
	Key        rune
}
