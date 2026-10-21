package interfaces

type IBaseware interface {
	RegisterUuid(uuid string) bool
	StartSakura(name string) error
	StopSakura(name string)
	GetGhosts() [][]string
	GetRunningGhosts() map[string]struct{}
	GetBalloons() [][]string
}
