package interfaces

import (
	"ninixw-core/ghost"
	"ninixw-core/structs"
)

type ISakuraInfo interface {
	GhostName() string
	GhostPath() string
	GhostRoot() string
	SakuraName() string
	KeroName() string
	LastScript() string
}

type IInstallObserver interface {
	SendEvent(event ...structs.Event) string
}

type IUpdateObserver interface {
	ISakuraInfo
	IInstallObserver
	ReloadShiori() error
	ReloadShell()
	ReloadBalloon()
}

type IBaseware interface {
	StartSakura(name string, reason string, from ISakuraInfo) error
	StopSakura(name string, uuid string)
	GetGhosts() [][]string
	GetRunningGhosts() map[string]struct{}
	GetBalloons() [][]string
	GetBalloon(b string) (ghost.Balloon, error)
	GetPreferences(s string) (string, bool)
	SetPreferences(k string, v string)
	InstallNar(f string, s IInstallObserver)
	UpdateGhost(url string, root string, s IUpdateObserver, reason string)
	OpenScriptLog()
	AppendScriptLog(kind string, data structs.LogData)
	OpenPreferences()
	NotifyConfigurationChanged()
	Add()
	Done()
}
