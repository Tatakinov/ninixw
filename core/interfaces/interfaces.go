package interfaces

type IBaseware interface {
	RegisterUuid(uuid string) bool
	StopSakura(name string)
	GetGhosts() [][]string
	GetRunningGhosts() map[string]struct{}
	GetBalloons() [][]string
}
