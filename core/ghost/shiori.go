package ghost

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ninixw-core/path"
	"ninixw-core/process"
)

type shiori struct {
	Ghost
	process process.Process
}

type Response struct {
	Method string
	Protocol string
	Header map[string]string
}

func newShiori(ghost Ghost) (shiori, error) {
	lib := ghost.Descript["shiori"]
	if lib == "" {
		return shiori{}, errors.New("shiori not found")
	}
	root := filepath.Join(path.GhostDir(), ghost.Key, "ghost", "master")
	info := getShioriFfiInfo(lib)
	lib = path.Lib(lib, root)
	args := make([]string, 0, len(info) + 2)
	args = append(args, "--run", lib, root)
	args = append(args, info...)
	s := []string{
		os.Getenv("NINIXW_PROXY_PATH"),
		path.ExecutableDir(),
	}
	for _, p := range s {
		process, err := process.NewProcess(filepath.Join(p, "ninixw-proxy.exe"), args...)
		if err != nil {
			continue
		}
		return shiori {
			Ghost: ghost,
			process: process,
		}, nil
	}
	return shiori{}, errors.New("shiori not found")
}

func (s shiori) Request(method string, command string, args ...string) (Response, bool) {
	req := make([]string, 0, len(args) + 5)
	req = append(req, fmt.Sprintf("%s SHIORI/3.0", method), "Charset: UTF-8", "Sender: ninixw", fmt.Sprintf("ID: %s", command))
	for i, arg := range args {
		if arg == "" {
			continue
		}
		req = append(req, fmt.Sprintf("Reference%d: %s", i, arg))
	}
	req = append(req, "\r\n")
	r := strings.Join(req, "\r\n")
	ch := make(chan string, 1)
	s.process.Send(method == "GET", ch, r)
	if method == "GET" {
		data, ok := <- ch
		if !ok {
			return Response{}, false
		}
		iss := strings.NewReader(data)
		reader := bufio.NewReaderSize(iss, len(data))
		bytes, _, err := reader.ReadLine()
		if err != nil {
			return Response{}, false
		}
		line := string(bytes)
		line = strings.TrimRight(line, "\r\n")
		s := strings.SplitN(line, " ", 2)
		if len(s) < 2 {
			return Response{}, false
		}
		m := s[0]
		p := s[1]
		h := make(map[string]string)
		for {
			bytes, _, err = reader.ReadLine()
			line = string(bytes)
			line = strings.TrimRight(line, "\r\n")
			if line == "" {
				break
			}
			s = strings.SplitN(line, ":", 2)
			if len(s) != 2 {
				return Response{}, false
			}
			key := strings.TrimSpace(s[0])
			value := strings.TrimSpace(s[1])
			h[key] = value
		}
		return Response {
			Method: m,
			Protocol: p,
			Header: h,
		}, true
	} else {
		return Response{
			Method: "NOTIFY",
			Protocol: "SHIORI/3.0",
			Header: make(map[string]string),
		}, true
	}
}

func (s shiori) Stop() error {
	return s.process.Stop()
}
