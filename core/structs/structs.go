package structs

type Event struct {
	Method string
	Option map[string]struct{}
	Event string
	Reference []string
}

