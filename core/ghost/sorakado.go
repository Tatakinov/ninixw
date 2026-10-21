package ghost

import (
	"fmt"
	"strings"

	"ninixw-core/process"
)

type sorakado struct {
	process process.Process
}

func (s sorakado) Request(method string, command string, args ...string) (string, bool) {
	req := make([]string, 0, len(args) + 4)
	req = append(req, fmt.Sprintf("%s SORAKADO/0.1", method), "Charset: UTF-8", fmt.Sprintf("Command: %s", command))
	for i, arg := range args {
		req = append(req, fmt.Sprintf("Argument%d: %s", i, arg))
	}
	req = append(req, "\r\n")
	r := strings.Join(req, "\r\n")
	ch := make(chan string, 1)
	s.process.Send(method == "GET", ch, r)
	if method == "GET" {
		data, ok := <- ch
		return data, ok
	} else {
		return "", false
	}
}

func (s sorakado) Stop() error {
	return s.process.Stop()
}
