package sakura

type MenuItem struct {
	Type string		`json:"type"`
	Caption string	`json:"caption"`
	Valid bool		`json:"valid"`
	Command string	`json:"command"`
	Args []string	`json:"args"`
	List []MenuItem	`json:"list"`
	State bool		`json:"state"`
}
