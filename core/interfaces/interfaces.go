package interfaces

import (
	"ninixw-core/ghost"
)

type ISakura interface {
	GhostPath() string
	GhostName() string
	SakuraName() string
	LastScript() string
}

type IBaseware interface {
	StartSakura(name string, reason string, from ISakura) error
	StopSakura(name string, uuid string)
	GetGhosts() [][]string
	GetRunningGhosts() map[string]struct{}
	GetBalloons() [][]string
	GetBalloon(b string) (ghost.Balloon, error)
	GetPreferences(s string) (string, bool)
	SetPreferences(k string, v string)
}
