package interfaces

import (
	"ninixw-core/ghost"
	"ninixw-core/structs"
)

type ISakura interface {
	SendEvent(event ...structs.Event)
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
	InstallNar(f string, s ISakura)
	UpdateGhost(url string, root string, s ISakura, reason string)
	OpenScriptLog()
	AppendScriptLog(kind string, data structs.LogData)
}
