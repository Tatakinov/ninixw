package sakura

import (
	"bufio"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"ninixw-core/ghost"
	"ninixw-core/log"
	"ninixw-core/path"
	"ninixw-core/process"
	"ninixw-core/utils"
)

type shiori struct {
	ghost.Ghost
	process process.Process
	charset string
}

type Response struct {
	Protocol string
	Code int
	Message string
	Header map[string]string
}

func newShiori(ghost ghost.Ghost) (shiori, error) {
	lib := ghost.Descript["shiori"]
	if lib == "" {
		lib = "shiori.dll"
	}
	root := filepath.Join(path.GhostDir(), ghost.Key, "ghost", "master")
	root = path.AppendDirSeparator(root)
	info := getShioriFfiInfo(lib)
	lib = path.Lib(lib, root)
	args := make([]string, 0, len(info) + 2)
	args = append(args, "--run", lib, root)
	args = append(args, info...)
	s := []string{
		path.ProxyDir(),
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
			charset: "",
		}, nil
	}
	return shiori{}, errors.New("shiori not found")
}

func (s shiori) Request(method string, command string, args ...string) (Response, error) {
	req := make([]string, 0, len(args) + 5)
	charset := "UTF-8"
	if s.charset != "" {
		charset = s.charset
	}
	req = append(req, fmt.Sprintf("%s SHIORI/3.0", method), fmt.Sprintf("Charset: %s", charset), "Sender: ninixw", fmt.Sprintf("ID: %s", command))
	for i, arg := range args {
		if arg == "" {
			continue
		}
		req = append(req, fmt.Sprintf("Reference%d: %s", i, arg))
	}
	req = append(req, "\r\n")
	r := strings.Join(req, "\r\n")
	switch charset {
	case "Shift_JIS":
		tmp, _, err := transform.String(japanese.ShiftJIS.NewEncoder(), r)
		if err != nil {
			log.Println(err)
			break
		}
		r = tmp
	}
	ch := make(chan string, 1)
	s.process.Send(method == "GET" || s.charset == "", ch, r)
	if method == "GET" || s.charset == "" {
		data, ok := <- ch
		if !ok {
			return Response{}, errors.New("channel closed")
		}
		f := func() (Response, error) {
			iss := strings.NewReader(data)
			reader := bufio.NewReader(iss)
			if s.charset == "Shift_JIS" {
				reader = bufio.NewReader(transform.NewReader(iss, japanese.ShiftJIS.NewDecoder()))
			}
			bytes, err := utils.ReadLine(reader)
			if err != nil {
				return Response{}, err
			}
			line := string(bytes)
			line = strings.TrimRight(line, "\r\n")
			slice := strings.SplitN(line, " ", 3)
			if len(slice) < 3 {
				return Response{}, errors.New("invalid shiori status line")
			}
			p := slice[0]
			c, err := strconv.Atoi(slice[1])
			if err != nil {
				return Response{}, err
			}
			m := slice[2]
			h := make(map[string]string)
			for {
				bytes, err = utils.ReadLine(reader)
				if err != nil {
					log.Println(err)
					return Response{}, err
				}
				line = string(bytes)
				line = strings.TrimRight(line, "\r\n")
				if line == "" {
					break
				}
				slice = strings.SplitN(line, ":", 2)
				if len(slice) != 2 {
					return Response{}, errors.New("invalid shiori header")
				}
				key := strings.TrimSpace(slice[0])
				value := strings.TrimSpace(slice[1])
				h[key] = value
			}
			return Response {
				Protocol: p,
				Code: c,
				Message: m,
				Header: h,
			}, nil
		}
		if s.charset == "" {
			res, err := f()
			if err != nil {
				return Response{}, err
			}
			switch res.Header["Charset"] {
			case "Shift_JIS", "":
				s.charset = "Shift_JIS"
			default:
				s.charset = "UTF-8"
			}
		}
		return f()
	} else {
		return Response{
			Protocol: "SHIORI/3.0",
			Code: 204,
			Message: "No Content",
			Header: make(map[string]string),
		}, nil
	}
}

func (s shiori) Stop() error {
	return s.process.Stop()
}
