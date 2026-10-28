package sakura

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"uuid"

	"ninixw-core/ghost"
	"ninixw-core/interfaces"
	"ninixw-core/log"
	"ninixw-core/path"
	"ninixw-core/sakurascript"
	"ninixw-core/sstp"
	"ninixw-core/structs"
	"ninixw-core/version"
	"ninixw-core/with_mutex"
)

const kBalloonTimeout = 15

type script struct {
	Data string
	Embed bool
}

type Rect struct {
	X int
	Y int
	W int
	H int
}

const kMonitor = "monitor"
const kSurface = "surface"
const kBalloon = "balloon"
const kBalloonOffset = "balloon_offset"

type rectKey struct {
	side int
	kind string
}

type rectInfo struct {
	m sync.RWMutex
	rect map[rectKey]Rect
}

type directionInfo struct {
	m sync.RWMutex
	direction map[int]bool
}

type Sakura struct {
	shiori shiori
	ao ao
	ai ai
	current_shell string
	current_balloon string
	parent interfaces.IBaseware
	alive bool
	quit chan struct{}
	script_ch chan script
	last_script string
	timer_ch with_mutex.WithMutex[string, func()]
	wait_timer *time.Timer
	balloon_timeout int
	quick_section bool
	paused bool
	uuid string
	uuid_ao string
	uuid_ai string
	side int
	sync_side []int
	rect with_mutex.WithMutex[rectKey, Rect]
	direction_info directionInfo
}

func toSide(side int) string {
	switch side {
	case 0:
		return "sakura"
	case 1:
		return "kero"
	default:
		return "char" + strconv.Itoa(side)
	}
}

func NewSakura(b interfaces.IBaseware, ghost ghost.Ghost, shell string, balloon ghost.Balloon, u string) (*Sakura, error) {
	shiori, err := newShiori(ghost)
	if err != nil {
		return &Sakura{}, err
	}
	ao, err := newAo(ghost.Shells[shell])
	if err != nil {
		shiori.process.Stop()
		return &Sakura{}, err
	}
	ai, err := newAi(balloon)
	if err != nil {
		shiori.process.Stop()
		ao.process.Stop()
		return &Sakura{}, err
	}
	u_ao := uuid.New().String()
	u_ai := uuid.New().String()
	s := &Sakura{
		shiori: shiori,
		ao: ao,
		ai: ai,
		current_shell: shell,
		current_balloon: balloon.Key,
		parent: b,
		alive: true,
		quit: make(chan struct{}),
		script_ch: make(chan script, 100),
		last_script: "",
		timer_ch: with_mutex.NewWithMutex[string, func()](),
		wait_timer: time.NewTimer(0 * time.Second),
		balloon_timeout: kBalloonTimeout,
		quick_section: false,
		paused: false,
		uuid: u,
		uuid_ao: u_ao,
		uuid_ai: u_ai,
		side: 0,
		sync_side: make([]int, 0),
		rect: with_mutex.NewWithMutex[rectKey, Rect](),
		direction_info: directionInfo{ m: sync.RWMutex{}, direction: make(map[int]bool) },
	}
	err = s.startServer()
	if err != nil {
		s.Stop()
		return &Sakura{}, err
	}
	return s, nil
}

func (s *Sakura) startServer() error {
	p := filepath.Join(path.SockDir(), s.uuid)
	server, err := net.Listen("unix", p)
	if err != nil {
		return err
	}
	go func() {
		<- s.quit
		server.Close()
		os.Remove(p)
	}()
	go sstp.HandleSSTP(s, server)
	return nil
}

func (s *Sakura) fromAo(req sstp.SstpRequest) bool {
	return s.uuid_ao == req.Header["Ao"]
}

func (s *Sakura) fromAi(req sstp.SstpRequest) bool {
	return s.uuid_ai == req.Header["Ai"]
}

func (s *Sakura) ProcessSSTP(r sstp.IRequest) sstp.Response {
	req, err := r.Sstp()
	if err != nil {
		log.Println(err)
		return sstp.NewResponse(400, "Bad Request", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
	}
	switch req.Method {
	case "NOTIFY", "SEND":
		event, _ := req.Header["Event"]
		if event != "" {
			script := s.SendEvent(structs.Event{ Method: "GET", Option: make(map[string]struct{}), Event: event, Reference: req.Reference })
			if script != "" {
				header := sstp.NewSSTPHeader(req.KeepAlive())
				header["Script"] = script
				return sstp.NewResponse(200, "OK", header, "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
			}
		} else {
			// TODO ifghost
			data, _ := req.Script[""]
			if data != "" {
				s.script_ch <- script{ Data: data, Embed: false }
				s.parent.AppendScriptLog("info", structs.LogData{
					Time: time.Now(),
					Ghost: s.shiori.Descript["name"],
					Status: make(map[string]struct{}),
					Script: data,
				})
			}
		}
	case "EXECUTE":
		command, ok := req.Header["Command"]
		if !ok {
			return sstp.NewResponse(400, "Bad Request", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
		}
		switch command {
		case "UpdateMonitorRect", "UpdateSurfaceRect", "UpdateBalloonRect":
			if len(req.Reference) != 5 {
				return sstp.NewResponse(400, "Bad Request", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
			}
			kind := kBalloon
			switch command {
			case "UpdateMonitorRect", "UpdateSurfaceRect":
				if !s.fromAo(req) {
					return sstp.NewResponse(400, "Bad Request", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
				}
				if command == "UpdateMonitorRect" {
					kind = kMonitor
				} else {
					kind = kSurface
				}
			case "UpdateBalloonRect":
				if !s.fromAi(req) {
					return sstp.NewResponse(400, "Bad Request", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
				}
			}
			side, err := strconv.Atoi(req.Reference[0])
			if err != nil {
				log.Println(err)
				return sstp.NewResponse(400, "Bad Request", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
			}
			x, err := strconv.Atoi(req.Reference[1])
			if err != nil {
				log.Println(err)
				return sstp.NewResponse(400, "Bad Request", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
			}
			y, err := strconv.Atoi(req.Reference[2])
			if err != nil {
				log.Println(err)
				return sstp.NewResponse(400, "Bad Request", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
			}
			w, err := strconv.Atoi(req.Reference[3])
			if err != nil {
				log.Println(err)
				return sstp.NewResponse(400, "Bad Request", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
			}
			h, err := strconv.Atoi(req.Reference[4])
			if err != nil {
				log.Println(err)
				return sstp.NewResponse(400, "Bad Request", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
			}
			s.rect.Lock(func(m map[rectKey]Rect) {
				m[rectKey{ side: side, kind: kind }] = Rect{ X: x, Y: y, W: w, H: h }
			})
		case "ResetBalloonPosition":
			if !s.fromAo(req) && !s.fromAi(req) {
				return sstp.NewResponse(400, "Bad Request", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
			}
			side, err := strconv.Atoi(req.Reference[0])
			if err != nil {
				log.Println(err)
				return sstp.NewResponse(400, "Bad Request", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
			}
			var ok bool
			var monitor Rect
			var surface Rect
			var balloon Rect
			var offset Rect
			s.rect.RLock(func(m map[rectKey]Rect) {
				monitor, ok = m[rectKey{ side: side, kind: kMonitor }]
				if !ok {
					return
				}
				surface, ok = m[rectKey{ side: side, kind: kSurface }]
				if !ok {
					return
				}
				balloon, ok = m[rectKey{ side: side, kind: kBalloon }]
				if !ok {
					return
				}
				// offsetは無くても良い
				offset, _ = m[rectKey{ side: side, kind: kBalloonOffset }]
			})
			if !ok {
				return sstp.NewResponse(204, "No Content", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
			}
			s.direction_info.m.RLock()
			direction, ok := s.direction_info.direction[side]
			s.direction_info.m.RUnlock()
			lx := surface.X - balloon.W + offset.X
			rx := surface.X + surface.W - offset.X
			x := 0
			if direction {
				if rx + balloon.W > monitor.X + monitor.W && lx >= monitor.X {
					direction = !direction
				}
			} else {
				if rx + balloon.W <= monitor.X + monitor.W && lx < monitor.X {
					direction = !direction
				}
			}
			if direction {
				if rx + balloon.W > monitor.X + monitor.W {
					rx = monitor.X + monitor.W - balloon.W
				}
				x = rx
			} else {
				if lx < monitor.X {
					lx = monitor.X
				}
				x = lx
			}
			y := surface.Y + offset.Y
			var i_direction string
			if direction {
				i_direction = "1"
			} else {
				i_direction = "0"
			}
			s.direction_info.m.Lock()
			s.direction_info.direction[side] = direction
			s.direction_info.m.Unlock()
			s.ai.Request("NOTIFY", "SetDirection", strconv.Itoa(side), i_direction)
			s.ai.Request("NOTIFY", "SetPosition", strconv.Itoa(side), strconv.Itoa(x - monitor.X), strconv.Itoa(y - monitor.Y), strconv.Itoa(monitor.X), strconv.Itoa(monitor.Y))
		case "AnalyzeFileMagic":
			if !s.fromAo(req) && !s.fromAi(req) {
				return sstp.NewResponse(400, "Bad Request", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
			}
			if len(req.Reference) <= 1 {
				return sstp.NewResponse(400, "Bad Request", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
			}
			mimes := make([]string, 0, len(req.Reference) - 1)
			for _, f := range req.Reference[1:] {
				file, err := os.Open(f)
				if err != nil {
					log.Println(err)
					continue
				}
				defer file.Close()
				bytes := make([]byte, 512)
				n, err := file.Read(bytes)
				if err != nil && err != io.EOF {
					log.Println(err)
					continue
				}
				mime := http.DetectContentType(bytes[:n])
				mimes = append(mimes, mime)
			}
			s.SendEvent(structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnFileDrop2",
				Reference: []string{
					strings.Join(req.Reference[1:], "\x01"),
					req.Reference[0],
					strings.Join(mimes, "\x01"),
				},
			})
			for i, m := range mimes {
				switch m {
				case "application/zip":
					s.parent.InstallNar(req.Reference[i + 1], s)
				default:
				}
			}
		case "OpenScriptLog":
			if !s.fromAo(req) {
				return sstp.NewResponse(400, "Bad Request", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
			}
			s.parent.OpenScriptLog()
		case "OpenPreferences":
			if !s.fromAo(req) {
				return sstp.NewResponse(400, "Bad Request", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
			}
			s.parent.OpenPreferences()
		case "UpdateGhost":
			if !s.fromAo(req) {
				return sstp.NewResponse(400, "Bad Request", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
			}
			homeurl := ""
			res, err := s.shiori.Request("GET", "homeurl")
			if err != nil {
				log.Println(err)
			} else {
				homeurl = res.Header["Value"]
			}
			if homeurl == "" {
				homeurl = s.shiori.Descript["homeurl"]
			}
			if strings.HasPrefix(homeurl, "http://") || strings.HasPrefix(homeurl, "https://") {
				root := s.shiori.Root
				s.parent.UpdateGhost(homeurl, root, s, "manual")
			}
		case "ChangeGhost":
			if !s.fromAo(req) {
				return sstp.NewResponse(400, "Bad Request", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
			}
			if len(req.Reference) == 0 {
				return sstp.NewResponse(400, "Bad Request", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
			}
			s.SendEvent(structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnGhostChanging",
				Reference: []string{ "user" },
			}, structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnClose",
				Reference: []string{ "user" },
			})
			<- s.quit
			err := s.parent.StartSakura(req.Reference[0], "change", s)
			if err != nil {
				log.Println(err)
			}
		case "SummonGhost":
			if !s.fromAo(req) {
				return sstp.NewResponse(400, "Bad Request", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
			}
			if len(req.Reference) == 0 {
				return sstp.NewResponse(400, "Bad Request", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
			}
			err := s.parent.StartSakura(req.Reference[0], "call", s)
			if err != nil {
				log.Println(err)
			}
		case "CloseGhost":
			s.SendEvent(structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnClose",
				Reference: []string{ "user" },
			})
		case "CloseAllGhost":
			s.SendEvent(structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnCloseAll",
				Reference: []string{ "user" },
			}, structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnClose",
				Reference: []string{ "user" },
			})
		}
	}
	return sstp.NewResponse(204, "No Content", sstp.NewSSTPHeader(req.KeepAlive()), "", sstp.Proto{ Name: "SSTP", Version: "1.4" })
}

func (s *Sakura) Stop() {
	close(s.quit)
	err := s.shiori.Stop()
	if err != nil {
		log.Println("shiori exited at", err)
	}
	err = s.ao.Stop()
	if err != nil {
		log.Println("ao exited at", err)
	}
	err = s.ai.Stop()
	if err != nil {
		log.Println("ai exited at", err)
	}
}

func (s *Sakura) MainLoop() {
	s.ao.Request("NOTIFY", "Initialize", path.AppendDirSeparator(s.ao.Path), "AO")
	s.ao.Request("NOTIFY", "BasewareVersion", "ninixw", version.Version)
	s.ao.Request("NOTIFY", "Endpoint", filepath.Join(path.SockDir(), s.uuid), s.uuid_ao)
	s.ai.Request("NOTIFY", "Initialize", path.AppendDirSeparator(s.ai.Path), "AI")
	s.ai.Request("NOTIFY", "BasewareVersion", "ninixw", version.Version)
	s.ai.Request("NOTIFY", "Endpoint", filepath.Join(path.SockDir(), s.uuid), s.uuid_ai)

	// \0のインスタンスはCreateが呼ばれずに処理が行われる可能性があるため
	// ここでCreateしておく
	s.ao.Request("NOTIFY", "Create", strconv.Itoa(s.side))
	s.ai.Request("NOTIFY", "Create", strconv.Itoa(s.side))

	go func () {
		defer s.Stop()
		defer s.wait_timer.Stop()
		asts := make([]sakurascript.Ast, 0)
		ch := make(chan struct{}, 100)
		defer close(ch)
		interpret_ticker := time.NewTicker(50 * time.Millisecond)
		defer interpret_ticker.Stop()
		onsecondchange_ticker := time.NewTicker(1 * time.Second)
		defer onsecondchange_ticker.Stop()
		timeout_ch := make(<- chan time.Time)
		restore_ch := make(<- chan time.Time)
		for {
			select {
			case data, ok := <- s.script_ch:
				if ok {
					s.last_script = data.Data
					tmp := sakurascript.Parse(sakurascript.Lex(data.Data))
					if len(tmp) == 0 {
						continue
					}
					ast := tmp[0]
					if ast.Kind == sakurascript.PTag && ast.Str == `\C` {
						for i := 0; i < len(asts); i++ {
							if asts[i].Kind == sakurascript.PTag && asts[i].Str == `\e` {
								asts = asts[:i]
								break
							}
						}
						asts = append(asts, tmp[1:]...)
					} else {
						asts = tmp
						// initialize
						s.quick_section = false
						s.side = 0
						s.ai.Request("NOTIFY", "HideAll")
						s.ai.Request("NOTIFY", "ClearTextAll")
						s.balloon_timeout = kBalloonTimeout
					}
					timeout_ch = make(<- chan time.Time)
					restore_ch = make(<- chan time.Time)
					interpret_ticker.Reset(50 * time.Millisecond)
					ch <- struct{}{}
				} else {
					break
				}
			case <- interpret_ticker.C:
				if !s.paused {
					ch <- struct{}{}
				}
			case <- onsecondchange_ticker.C:
				if len(asts) == 0 {
					s.SendEvent(structs.Event{ Method: "GET", Option: make(map[string]struct{}), Event: "OnSecondChange", Reference: []string{ "", "0", "0", "1" } })
				}
			case <- timeout_ch:
				timeout_ch = make(chan time.Time)
				s.ai.Request("NOTIFY", "HideAll")
				restore_ch = time.After(15 * time.Second)
			case <- restore_ch:
				restore_ch = make(chan time.Time)
				// TODO get surface id
				s.SendEvent(structs.Event{ Method: "GET", Option: make(map[string]struct{}), Event: "OnSurfaceRestore", Reference: []string{} })
			case <- s.wait_timer.C:
				s.paused = false
				ch <- struct{}{}
			case <- ch:
				if len(asts) == 0 && !s.alive {
					return
				}
				if len(asts) == 0 || s.paused {
					continue
				}
				for {
					tmp, err := s.interpretScript(asts)
					if err != nil {
						log.Println(err)
					}
					asts = tmp
					if !s.quick_section || s.paused || len(asts) == 0 {
						break
					}
				}
				if len(asts) == 0 {
					if s.balloon_timeout > 0 {
						timeout_ch = time.After(time.Duration(s.balloon_timeout) * time.Second)
					}
				}
			}
		}
	}()
	<- s.quit
	s.parent.StopSakura(s.shiori.Key, s.uuid)
}

func (s *Sakura) interpretScript(asts []sakurascript.Ast) ([]sakurascript.Ast, error) {
	index := 0
	in_q := false
	in_a := false
	for {
		if index >= len(asts) {
			break
		}
		ast := asts[index]
		index++
		switch ast.Kind {
		case sakurascript.PString:
			slice := strings.SplitN(ast.Str, "", 2)
			s.ai.Request("NOTIFY", "AppendText", strconv.Itoa(s.side), slice[0])
			if len(slice) == 2 {
				index--
				ast.Str = slice[1]
				asts[index] = ast
				if !in_q && !in_a {
					return asts[index:], nil
				}
			} else {
				if !in_q && !in_a {
					return asts[index:], nil
				}
			}
		case sakurascript.PTag:
			switch ast.Str {
			case `\-`:
				s.alive = false
				return make([]sakurascript.Ast, 0), nil
			case `\!`:
				if len(ast.Args) == 0 {
					return asts[index:], errors.New(`invalid args size in \!`)
				}
				switch ast.Args[0] {
				case "raise", "notify":
					if len(ast.Args) < 2 {
						return asts[index:], errors.New(`invalid args size in \![raise/notify]`)
					}
					method := "GET"
					if ast.Args[0] == "notify" {
						method = "NOTIFY"
					}
					s.SendEvent(structs.Event{ Method: method, Option: make(map[string]struct{}), Event: ast.Args[1], Reference: ast.Args[2:] })
				case "timerraise", "timernotify":
					if len(ast.Args) < 4 {
						return asts[index:], errors.New(`invalid args size in \![raise/notify]`)
					}
					method := "GET"
					if ast.Args[0] == "timernotify" {
						method = "NOTIFY"
					}
					limit, err := strconv.Atoi(ast.Args[1])
					if err != nil {
						return asts[index:], err
					}
					loop := false
					if ast.Args[2] == "0" {
						loop = true
					}
					if limit > 0 {
						var fn func()
						var ch <- chan time.Time
						if loop {
							t := time.NewTicker(time.Duration(limit) * time.Millisecond)
							fn = func() {
								t.Stop()
							}
							ch = t.C
						} else {
							t := time.NewTimer(time.Duration(limit) * time.Millisecond)
							fn = func() {
								t.Stop()
							}
							ch = t.C
						}
						s.timer_ch.Lock(func(m map[string]func()) {
							f, ok := m[ast.Args[3]]
							if ok {
								f()
							}
							m[ast.Args[3]] = fn
						})
						go func() {
							for {
								select {
								case <- s.quit:
									return
								case <- ch:
									s.SendEvent(structs.Event{ Method: method, Option: make(map[string]struct{}), Event: ast.Args[3], Reference: ast.Args[4:] })
									if !loop {
										s.timer_ch.Lock(func(m map[string]func()) {
											_, ok := m[ast.Args[3]]
											if ok {
												delete(m, ast.Args[3])
											}
										})
										break
									}
								}
							}
						}()
					} else {
						s.timer_ch.Lock(func(m map[string]func()) {
							f, ok := m[ast.Args[3]]
							if ok {
								f()
								delete(m, ast.Args[3])
							}
						})
					}
				case "bind":
					if len(ast.Args) != 4 {
						return asts[index:], errors.New(`invalid args size in \![bind]`)
					}
					flag := "toggle"
					if ast.Args[3] == "1" {
						flag = "true"
					} else if ast.Args[3] == "0" {
						flag = "false"
					}
					s.ao.Request("NOTIFY", "Bind", strconv.Itoa(s.side), ast.Args[1], ast.Args[2], "script", flag)
				default:
				}
			case `\&`:
			case `\*`:
			case `\+`:
			case `\0`, `\h`:
				if s.side == 0 {
					continue
				}
				s.side = 0
				s.ao.Request("NOTIFY", "Create", strconv.Itoa(s.side))
				s.ai.Request("NOTIFY", "Create", strconv.Itoa(s.side))
			case `\1`, `\u`:
				if s.side == 1 {
					continue
				}
				s.side = 1
				s.ao.Request("NOTIFY", "Create", strconv.Itoa(s.side))
				s.ai.Request("NOTIFY", "Create", strconv.Itoa(s.side))
			case `\4`:
			case `\5`:
			case `\6`:
			case `\7`:
			case `\8`:
			case `\C`:
				// nop
			case `\a`:
			case `\b`:
				if len(ast.Args) != 1 {
					return asts[index:], errors.New(`invalid args size in \b`)
				}
				n, err := strconv.Atoi(ast.Args[0])
				if err != nil {
					return asts[index:], err
				}
				if n == -1 {
					s.ai.Request("NOTIFY", "Hide", strconv.Itoa(s.side))
				} else if n >= 0 {
					s.ai.Request("NOTIFY", "SetBalloonID", strconv.Itoa(s.side), strconv.Itoa(n))
				}
			case `\c`:
				if len(ast.Args) == 0 {
					s.ai.Request("NOTIFY", "ClearText", strconv.Itoa(s.side))
				} else {
					// TODO stub
				}
			case `\e`, `\z`:
				return make([]sakurascript.Ast, 0), nil
			case `\f`:
				// TODO stub
			case `\i`:
				// TODO stub
			case `\j`:
				// TODO stub
			case `\m`:
				// TODO stub
			case `\n`:
				switch len(ast.Args) {
				case 0:
					s.ai.Request("NOTIFY", "NewLine", strconv.Itoa(s.side))
					s.ai.Request("NOTIFY", "SetCursorPosition", strconv.Itoa(s.side), "x", "0", "true", "px")
					s.ai.Request("NOTIFY", "SetCursorPosition", strconv.Itoa(s.side), "y", "1", "false", "lh")
				case 1:
					if ast.Args[0] == "half" {
						s.ai.Request("NOTIFY", "NewLine", strconv.Itoa(s.side))
						s.ai.Request("NOTIFY", "SetCursorPosition", strconv.Itoa(s.side), "x", "0", "true", "px")
						s.ai.Request("NOTIFY", "SetCursorPosition", strconv.Itoa(s.side), "y", "0.5", "false", "lh")
					} else {
						// TODO stub
					}
				default:
				}
			case `\p`:
				if len(ast.Args) != 1 {
					return asts[index:], errors.New(`invalid args size in \p`)
				}
				n, err := strconv.Atoi(ast.Args[0])
				if err != nil {
					return asts[index:], err
				}
				if s.side == n {
					continue
				}
				s.side = n
				s.ao.Request("NOTIFY", "Create", strconv.Itoa(s.side))
				s.ai.Request("NOTIFY", "Create", strconv.Itoa(s.side))
			case `\q`:
				if len(ast.Args) < 2 {
					return asts[index:], errors.New(`invalid args size in \q`)
				}
				side := strconv.Itoa(s.side)
				args := make([]string, 0, 3 + len(ast.Args) - 1)
				args = append(args, side)
				args = append(args, "false")
				args = append(args, ast.Args[1:]...)
				s.ai.Request("NOTIFY", "AppendLinkBegin", args...)
				s.ai.Request("NOTIFY", "AppendText", side, ast.Args[0])
				s.ai.Request("NOTIFY", "AppendLinkEnd", side)
			case `\s`:
				if len(ast.Args) != 1 {
					return asts[index:], errors.New(`invalid args size in \s`)
				}
				if ast.Args[0] == "-1" {
					s.ao.Request("NOTIFY", "Hide", strconv.Itoa(s.side))
				} else {
					s.ao.Request("NOTIFY", "SetSurfaceID", strconv.Itoa(s.side), ast.Args[0])
				}
			case `\t`:
				// TODO stub
			case `\v`:
				// TODO stub
			case `\w`:
				if len(ast.Args) != 1 {
					return asts[index:], errors.New(`invalid args size in \w`)
				}
				n, err := strconv.Atoi(ast.Args[0])
				if err != nil {
					return asts[index:], err
				}
				s.paused = true
				s.wait_timer.Reset(time.Duration(n * 50) * time.Millisecond)
				return asts[index:], nil
			case `\x`:
				// TODO stub
			case `\_!`, `\_?`:
				// parseの段階で処理が終わっているのでnop
			case `\_V`:
				// TODO stub
			case `\_a`:
				// TODO stub
			case `\_b`:
				if len(ast.Args) < 2 {
					return asts[index:], errors.New(`invalid args size \_b`)
				}
				p := ast.Args[0]
				if !filepath.IsAbs(p) {
					p = filepath.Join(s.shiori.Path, p)
				}
				if !strings.HasPrefix(p, path.HomeDir()) {
					log.Println("ninixw", "directory traversal?", p)
				}
				i := 1
				x := "0"
				y := "0"
				inline := "false"
				if ast.Args[i] == "inline" {
					inline = "true"
					i++
				} else {
					if len(ast.Args) < 3 {
						return asts[index:], errors.New(`invalid args size \_b`)
					}
					_, err := strconv.Atoi(ast.Args[i])
					if err != nil {
						return asts[index:], err
					}
					x = ast.Args[i]
					i++
					_, err = strconv.Atoi(ast.Args[i])
					if err != nil {
						return asts[index:], err
					}
					y = ast.Args[i]
					i++
				}
				opaque := "false"
				use_self_alpha := "false"
				clipping := ""
				scaling := ""
				fixed := "false"
				foreground := "false"
				source := ""
				for ; i < len(ast.Args); i++ {
					a := ast.Args[i]
					switch a {
					case "opaque", "--option=opaque":
						opaque = "true"
					case "--option=use_self_alpha":
						use_self_alpha = "true"
					case "--option=fixed":
						fixed = "true"
					case "--option=background":
						foreground = "false"
					case "--option=foreground":
						foreground = "true"
					default:
						if strings.HasPrefix(a, "--clipping=") {
							clipping = a[11:]
						} else if strings.HasPrefix(a, "--scaling=") {
							scaling = a[10:]
						} else if strings.HasPrefix(a, "--source=") {
							source = a[9:]
						}
					}
				}
				s.ai.Request("NOTIFY", "AppendImage", strconv.Itoa(s.side), p, x, y, inline, opaque, use_self_alpha, clipping, scaling, fixed, foreground, source)
			case `\_l`:
				if len(ast.Args) != 2 {
					return asts[index:], errors.New(`invalid args size \_l`)
				}
				if ast.Args[0] == "" && ast.Args[1] == "" {
					continue
				}
				s.ai.Request("NOTIFY", "NewLine", strconv.Itoa(s.side))
				axis := []string{ "x", "y" }
				units := []string{ "px", "em", "lh", "%" }
				for i := 0; i < 2; i++ {
					arg := ast.Args[i]
					if arg == "" {
						continue
					}
					if arg == "center" + axis[i] {
						// TODO stub
					}
					is_absolute := "true"
					if strings.HasPrefix(arg, "@") {
						is_absolute = "false"
						arg = arg[1:]
					}
					unit := "px"
					for _, u := range units {
						if strings.HasSuffix(arg, u) {
							unit = u
							break
						}
					}
					value, err := strconv.ParseFloat(arg, 64)
					if err != nil {
						return asts[index:], err
					}
					if unit == "%" {
						unit = "em"
						value /= 100.0
					}
					s.ai.Request("NOTIFY", "SetCursorPosition", strconv.Itoa(s.side), axis[i], strconv.FormatFloat(value, 'f', -1, 64), is_absolute, unit)
				}
			case `\_m`:
				// TODO stub
			case `\_n`:
				// TODO stub
			case `\_q`:
				s.quick_section = !s.quick_section
			case `\_s`:
				// TODO stub
			case `\_u`:
				// TODO stub
			case `\_v`:
				// TODO stub
			case `\_w`:
				if len(ast.Args) != 1 {
					return asts[index:], errors.New(`invalid args size in \_w`)
				}
				n, err := strconv.Atoi(ast.Args[0])
				if err != nil {
					return asts[index:], err
				}
				s.paused = true
				s.wait_timer.Reset(time.Duration(n) * time.Millisecond)
				return asts[index:], nil
			case `\__c`:
				// TODO stub
			case `\__q`:
				if in_q && len(ast.Args) == 0 {
					s.ai.Request("NOTIFY", "AppendLinkEnd", strconv.Itoa(s.side))
					in_q = false
					return asts[index:], nil
				} else if !in_q && len(ast.Args) > 0 {
					in_q = true
					args := make([]string, 0, 2 + len(ast.Args))
					args = append(args, strconv.Itoa(s.side))
					args = append(args, "false")
					args = append(args, ast.Args...)
					s.ai.Request("NOTIFY", "AppendLinkBegin", args...)
				}
			case `\__t`:
				// TODO stub
			case `\__v`:
				// TODO stub
			case `\__w`:
				// TODO stub
			default:
			}
		case sakurascript.PEnv:
			switch ast.Str {
			default:
			}
		}
	}
	return asts[index:], nil
}

func (s *Sakura) SendEvent(event ...structs.Event) string {
	if len(event) == 0 {
		return ""
	}
	select {
	case <- s.quit:
		return ""
	default:
		// nop
	}
	var e structs.Event
	v := ""
	for _, e = range event {
		res, err := s.shiori.Request(e.Method, e.Event, e.Reference...)
		if err != nil {
			log.Println(err)
			return ""
		}
		tmp, ok := res.Header["Value"]
		if ok {
			v = tmp
			break
		}
	}
	if v != "" {
		if _, ok := e.Option["notranslate"]; ok {
			_, ok = e.Option["embed"]
			s.script_ch <- script{ Data: v, Embed: ok }
			s.parent.AppendScriptLog("info", structs.LogData{
				Time: time.Now(),
				Ghost: s.shiori.Descript["name"],
				Status: e.Option,
				Script: v,
			})
		} else {
			res, err := s.shiori.Request("GET", "OnTranslate", v, "", e.Event, strings.Join(e.Reference, "\x01"))
			if err != nil {
				log.Println(err)
			}
			t, ok := res.Header["Value"]
			if ok {
				_, ok = e.Option["embed"]
				s.script_ch <- script{ Data: t, Embed: ok }
				s.parent.AppendScriptLog("info", structs.LogData{
					Time: time.Now(),
					Ghost: s.shiori.Descript["name"],
					Status: e.Option,
					Script: t,
				})
				v = t
			} else {
				_, ok = e.Option["embed"]
				s.script_ch <- script{ Data: v, Embed: ok }
				s.parent.AppendScriptLog("info", structs.LogData{
					Time: time.Now(),
					Ghost: s.shiori.Descript["name"],
					Status: e.Option,
					Script: v,
				})
			}
		}
	}
	switch e.Event {
	case "OnCloseAll", "OnClose", "OnGhostChanging":
		s.alive = false
	case "OnMouseClick":
		if len(e.Reference) < 6 || e.Reference[5] != "1" || v != "" {
			return v
		}
		s.openMenu(e.Reference[3])
	}
	return v
}

func (s *Sakura) openMenu(from string) {
	n, err := strconv.Atoi(from)
	if err != nil {
		log.Println(err)
		return
	}
	menu := make([]MenuItem, 0)
	side := toSide(n)
	res, err := s.shiori.Request("GET", side + ".portalsites")
	if err != nil {
		log.Println(err)
		return
	}
	submenu := make([]MenuItem, 0)
	for _, line := range strings.Split(res.Header["Value"], "\x02") {
		slice := strings.Split(line, "\x01")
		if len(slice) < 2 {
			continue
		}
		submenu = append(submenu, MenuItem{
			Type: "item",
			Caption: slice[0],
			Valid: true,
			Command: "VisitSite",
			Args: slice[1:],
		})
	}
	menu = append(menu, MenuItem{
		Type: "submenu",
		Caption: "portal",
		List: submenu,
	})
	res, err = s.shiori.Request("GET", side + ".recommendsites")
	if err != nil {
		log.Println(err)
		return
	}
	submenu = make([]MenuItem, 0)
	for _, line := range strings.Split(res.Header["Value"], "\x02") {
		slice := strings.Split(line, "\x01")
		if len(slice) < 2 {
			continue
		}
		submenu = append(submenu, MenuItem{
			Type: "item",
			Caption: slice[0],
			Valid: true,
			Command: "VisitSite",
			Args: slice[1:],
		})
	}
	menu = append(menu, MenuItem{
		Type: "submenu",
		Caption: "recommend",
		Valid: true,
		List: submenu,
	})
	homeurl := s.shiori.Descript["homeurl"]
	if homeurl == "" {
		res, err := s.shiori.Request("GET", "homeurl")
		if err != nil {
			log.Println(err)
		} else {
			homeurl = res.Header["Value"]
		}
	}
	if !strings.HasPrefix(homeurl, "http://") && strings.HasPrefix(homeurl, "https://") {
		homeurl = ""
	}
	menu = append(menu, MenuItem{
		Type: "item",
		Caption: "update",
		Valid: homeurl != "",
		Command: "UpdateGhost",
	})
	menu = append(menu, MenuItem{
		Type: "item",
		Caption: "vanish",
		Command: "VanishGhost",
	})
	menu = append(menu, MenuItem{
		Type: "submenu",
		Caption: "function",
		Valid: true,
		List: []MenuItem{
			MenuItem{
				Type: "item",
				Caption: "Script Inputbox",
				Valid: true,
				Command: "OpenScriptInputbox",
			},
			MenuItem{
				Type: "item",
				Caption: "Script Log",
				Valid: true,
				Command: "OpenScriptLog",
			},
		},
	})
	menu = append(menu, MenuItem{
		Type: "submenu",
		Caption: "preferences",
		Valid: true,
		List: []MenuItem{
			MenuItem{
				Type: "item",
				Caption: "Preferences",
				Valid: true,
				Command: "OpenPreferences",
			},
		},
	})
	ghosts := s.parent.GetGhosts()
	running := s.parent.GetRunningGhosts()
	l := make([]MenuItem, 0, len(ghosts))
	for _, v := range ghosts {
		_, ok := running[v[1]]
		l = append(l, MenuItem{
			Type: "item",
			Caption: v[0],
			Valid: !ok,
			Command: "ChangeGhost",
			Args: []string{ v[1] },
		})
	}
	menu = append(menu, MenuItem{
		Type: "submenu",
		Caption: "switch",
		Valid: true,
		List: l,
	})
	l = make([]MenuItem, 0, len(ghosts))
	for _, v := range ghosts {
		_, ok := running[v[1]]
		l = append(l, MenuItem{
			Type: "item",
			Caption: v[0],
			Valid: !ok,
			Command: "SummonGhost",
			Args: []string{ v[1] },
		})
	}
	menu = append(menu, MenuItem{
		Type: "submenu",
		Caption: "call",
		Valid: true,
		List: l,
	})
	shells := make([][]string, 0, len(s.shiori.Shells))
	for k, v := range s.shiori.Shells {
		shells = append(shells, []string{ v.Descript["name"], k})
	}
	slices.SortFunc(shells, func(a []string, b []string) int {
		result := cmp.Compare(a[0], b[0])
		if result == 0 {
			return cmp.Compare(a[1], b[1])
		}
		return result
	})
	l = make([]MenuItem, 0, len(shells))
	for _, v := range shells {
		l = append(l, MenuItem{
			Type: "item",
			Caption: v[0],
			Valid: s.current_shell != v[1],
			Command: "ChangeShell",
			Args: []string{ v[1] },
		})
	}
	menu = append(menu, MenuItem{
		Type: "submenu",
		Caption: "shell",
		Valid: true,
		List: l,
	})
	menu = append(menu, MenuItem{
		Type: "dressup",
		Caption: "costume",
		Valid: true,
		List: []MenuItem{},
	})
	balloons := s.parent.GetBalloons()
	l = make([]MenuItem, 0, len(balloons))
	for _, v := range balloons {
		l = append(l, MenuItem{
			Type: "item",
			Caption: v[0],
			Valid: s.current_balloon != v[1],
			Command: "ChangeBalloon",
			Args: []string{ v[1] },
		})
	}
	menu = append(menu, MenuItem{
		Type: "submenu",
		Caption: "balloon",
		Valid: true,
		List: l,
	})
	menu = append(menu, MenuItem{
		Type: "submenu",
		Caption: "info",
		Valid: true,
		List: []MenuItem{
			MenuItem{
				Type: "item",
				Caption: "ベースウェアについて",
				Valid: true,
				Command: "ShowBasewareVersion",
			},
		},
	})
	menu = append(menu, MenuItem{
		Type: "item",
		Caption: "close",
		Valid: true,
		Command: "CloseGhost",
	})
	menu = append(menu, MenuItem{
		Type: "item",
		Caption: "close all",
		Valid: true,
		Command: "CloseAllGhost",
	})
	bytes, err := json.Marshal(menu)
	if err != nil {
		log.Println(err)
		return
	}
	s.ao.Request("NOTIFY", "NotifyMenuInfo", string(bytes))
}

func (s *Sakura) Uuid() string {
	return s.uuid
}

func (s *Sakura) GhostName() string {
	return s.shiori.Descript["name"]
}

func (s *Sakura) GhostPath() string {
	return s.shiori.Path
}

func (s *Sakura) GhostRoot() string {
	return s.shiori.Root
}

func (s *Sakura) SakuraName() string {
	return s.shiori.Descript["sakura.name"]
}

func (s *Sakura) KeroName() string {
	return s.shiori.Descript["kero.name"]
}

func (s *Sakura) LastScript() string {
	return s.last_script
}

func (s *Sakura) ReloadShiori() error {
	s.shiori.Stop()
	sh, err := newShiori(ghost.Ghost{ Key: s.shiori.Key, Path: s.shiori.Path, Descript: s.shiori.Descript })
	if err != nil {
		return err
	}
	s.shiori = sh
	return nil
}

func (s *Sakura) LoadShell(shell string) error {
	s.ao.Stop()
	a, err := newAo(s.shiori.Shells[shell])
	if err != nil {
		return err
	}
	s.ao = a
	return nil
}

func (s *Sakura) ReloadShell() {
	s.ao.Request("NOTIFY", "Reload")
}

func (s *Sakura) LoadBalloon(b string) error {
	balloon, err := s.parent.GetBalloon(b)
	if err != nil {
		return err
	}
	s.ai.Stop()
	a, err := newAi(balloon)
	if err != nil {
		return err
	}
	s.ai = a
	return nil
}

func (s *Sakura) ReloadBalloon() {
	s.ai.Request("NOTIFY", "Reload")
}

func (s *Sakura) OnConfigurationChanged() {
	m := make(map[string]string)
	shell_scale, ok := s.parent.GetPreferences("shell_scale")
	if !ok {
		shell_scale = "100"
	}
	m["scale"] = shell_scale
	font_name, ok := s.parent.GetPreferences("font_name")
	if ok {
		m["font"] = font_name
	}
	args := make([]string, 0, len(m))
	for k, v := range m {
		args = append(args, fmt.Sprintf("%s,%s", k, v))
	}
	s.ao.Request("NOTIFY", "ConfigurationChanged", args...)
	s.ai.Request("NOTIFY", "ConfigurationChanged", args...)
}
