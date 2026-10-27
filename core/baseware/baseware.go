package baseware

import (
	"bufio"
	"cmp"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"uuid"

	"ninixw-core/ghost"
	"ninixw-core/interfaces"
	"ninixw-core/log"
	"ninixw-core/path"
	"ninixw-core/sakura"
	"ninixw-core/structs"
	"ninixw-core/utils"
	"ninixw-core/with_mutex"
)

type Baseware struct {
	ghosts with_mutex.WithMutex[string, ghost.Ghost]
	balloons with_mutex.WithMutex[string, ghost.Balloon]
	sakuras with_mutex.WithMutex[string, *sakura.Sakura]
	preferences with_mutex.WithMutex[string, string]
	wg sync.WaitGroup
	channel chan string
	uuids with_mutex.WithMutex[string, struct{}]
	installs with_mutex.WithMutex[string, struct{}]
	scriptlogProcess *scriptlog
	preferencesProcess *preferences
}

func NewBaseware() *Baseware {
	entries, err := os.ReadDir(path.SockDir())
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		os.Remove(filepath.Join(path.SockDir(), entry.Name()))
	}

	return &Baseware{
		ghosts: with_mutex.NewWithMutex[string, ghost.Ghost](),
		balloons: with_mutex.NewWithMutex[string, ghost.Balloon](),
		sakuras: with_mutex.NewWithMutex[string, *sakura.Sakura](),
		preferences: with_mutex.NewWithMutex[string, string](),
		wg: sync.WaitGroup{},
		channel: make(chan string),
		uuids: with_mutex.NewWithMutex[string, struct{}](),
		installs: with_mutex.NewWithMutex[string, struct{}](),
		scriptlogProcess: newScriptlog(),
		preferencesProcess: newPreferences(),
	}
}

func (b *Baseware) Load() {
	fmt.Println("Loading...")
	b.loadPreferences()
	dirs := path.Dirs(path.GhostDir())
	sum := 0
	for _, dir := range dirs {
		g, err := ghost.NewGhost(filepath.Join(path.GhostDir(), dir))
		if err != nil {
			continue
		}
		b.ghosts.Lock(func(m map[string]ghost.Ghost) {
			m[g.Root] = g
		})
		sum++
	}
	fmt.Println(fmt.Sprintf("Loading %d ghosts", sum))
	sum = 0
	dirs = path.Dirs(path.BalloonDir())
	for _, dir := range dirs {
		balloon, err := ghost.NewBalloon(filepath.Join(path.BalloonDir(), dir))
		if err != nil {
			continue
		}
		b.balloons.Lock(func(m map[string]ghost.Balloon) {
			m[dir] = balloon
		})
		sum++
	}
	fmt.Println(fmt.Sprintf("Loading %d balloons", sum))
}

func (b *Baseware) loadPreferences() {
	file, err := os.Open(filepath.Join(path.HomeDir(), "preferences.txt"))
	if err != nil {
		return
	}
	defer file.Close()

	r := bufio.NewReader(file)
	p := make(map[string]string)
	for {
		bytes, err := utils.ReadLine(r)
		if err != nil {
			break
		}
		line := string(bytes)
		line = strings.TrimRight(line, "\r\n")
		s := strings.SplitN(line, "=", 2)
		if len(s) < 2 {
			continue
		}
		p[strings.TrimSpace(s[0])] = strings.TrimSpace(s[1])
	}
	b.preferences.Lock(func(m map[string]string) {
		for k, v := range p {
			m[k] = v
		}
	})
}

func (b *Baseware) savePreferences() {
	p := make(map[string]string)
	b.preferences.RLock(func(m map[string]string) {
		for k, v := range m {
			p[k] = v
		}
	})
	s := make([]string, 0, len(p))
	for k, v := range p {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	err := os.WriteFile(filepath.Join(path.HomeDir(), "preferences.txt"), []byte(strings.Join(s, "\r\n")), 0644)
	if err != nil {
		log.Println(err)
	}
}

func (b *Baseware) Run(ghost_name string) {
	i := 1
	b.ghosts.RLock(func(m map[string]ghost.Ghost) {
		i *= len(m)
	})
	b.balloons.RLock(func(m map[string]ghost.Balloon) {
		i *= len(m)
	})
	if i == 0 {
		panic("no ghost or no balloon")
	} else {
		if ghost_name == "" {
			b.preferences.RLock(func(m map[string]string) {
				ghost_name = m["last_ghost"]
			})
			if ghost_name == "" {
				ghost_name = "knowledgeable_kernel"
			}
		}
		b.ghosts.RLock(func(m map[string]ghost.Ghost) {
			_, ok := m[ghost_name]
			if ok {
				return
			}
			ghost_name = filepath.Base(ghost_name)
			for k, _ := range m {
				name := filepath.Base(k)
				if ghost_name == name {
					ghost_name = k
					ok = true
					break
				}
			}
			if ok {
				return
			}
			for k, _ := range m {
				ghost_name = k
				break
			}
		})
		err := b.StartSakura(ghost_name, "system", &sakura.Sakura{})
		if err != nil {
			log.Println(err)
		}
	}
	b.wg.Wait()
	b.savePreferences()
}

func (b *Baseware) StartSakura(name string, reason string, from interfaces.ISakura) error {
	var g ghost.Ghost
	var balloon ghost.Balloon
	ok := false
	b.ghosts.RLock(func(m map[string]ghost.Ghost) {
		g, ok = m[name]
	})
	if !ok {
		b.ghosts.RLock(func(m map[string]ghost.Ghost) {
			for _, v := range m {
				if v.Descript["name"] == name {
					g = v
					ok = true
					break
				}
			}
		})
	}
	if !ok {
		return errors.New(fmt.Sprintf("ghost: %s not found", name))
	}
	s := "master"
	_, ok = g.Shells[s]
	if !ok {
		for k, _ := range g.Shells {
			s = k
			ok = true
			break
		}
	}
	if !ok {
		return errors.New(fmt.Sprintf("shell not found"))
	}
	b.balloons.RLock(func(m map[string]ghost.Balloon) {
		balloon, ok = m["balloon_terminal"]
		if ok {
			return
		}
		for _, v := range m {
			balloon = v
			ok = true
			break
		}
	})
	if !ok {
		return errors.New(fmt.Sprintf("shell not found"))
	}
	u := ""
	b.uuids.Lock(func(m map[string]struct{}) {
		for {
			u = uuid.New().String()
			_, ok := m[u]
			if ok {
				continue
			}
			m[u] = struct{}{}
			break
		}
	})
	instance, err := sakura.NewSakura(b, g, s, balloon, u)
	if err != nil {
		return err
	}
	b.sakuras.Lock(func(m map[string]*sakura.Sakura) {
		m[g.Root] = instance
	})
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		defer func() {
			b.preferences.Lock(func (m map[string]string) {
				m["last_ghost"] = g.Root
			})
		}()
		instance.SendEvent(structs.Event{
			Method: "NOTIFY",
			Option: make(map[string]struct{}),
			Event: "OnInitialize",
			Reference: []string{ },
		})
		switch reason {
		case "system":
			instance.SendEvent(structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnBoot",
				Reference: []string{ g.Shells[s].Descript["name"] },
			})
		case "change":
			instance.SendEvent(structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnGhostChanged",
				Reference: []string{ from.SakuraName(), from.LastScript(), from.GhostName(), from.GhostPath() },
			}, structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnBoot",
				Reference: []string{ g.Shells[s].Descript["name"] },
			})
		case "call":
			instance.SendEvent(structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnGhostCalled",
				Reference: []string{ from.SakuraName(), from.LastScript(), from.GhostName(), from.GhostPath() },
			}, structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnBoot",
				Reference: []string{ g.Shells[s].Descript["name"] },
			})
		}
		instance.MainLoop()
	}()
	return nil
}

func (b *Baseware) StopSakura(name string, uuid string) {
	b.sakuras.Lock(func(m map[string]*sakura.Sakura) {
		delete(m, name)
	})
	b.uuids.Lock(func(m map[string]struct{}) {
		delete(m, uuid)
	})
}

func (b *Baseware) GetGhosts() [][]string {
	s := make([][]string, 0)
	b.ghosts.Lock(func(m map[string]ghost.Ghost) {
		for k, v := range m {
			s = append(s, []string{ v.Descript["name"], k })
		}
	})
	slices.SortFunc(s, func(a []string, b []string) int {
		result := cmp.Compare(a[0], b[0])
		if result == 0 {
			return cmp.Compare(a[1], b[1])
		}
		return result
	})
	return s
}

func (b *Baseware) GetRunningGhosts() map[string]struct{} {
	running_ghosts := make(map[string]struct{})
	b.sakuras.Lock(func(m map[string]*sakura.Sakura) {
		for k, _ := range m {
			running_ghosts[k] = struct{}{}
		}
	})
	return running_ghosts
}

func (b *Baseware) GetBalloons() [][]string {
	s := make([][]string, 0)
	b.balloons.Lock(func(m map[string]ghost.Balloon) {
		for k, v := range m {
			s = append(s, []string{ v.Descript["name"], k })
		}
	})
	slices.SortFunc(s, func(a []string, b []string) int {
		result := cmp.Compare(a[0], b[0])
		if result == 0 {
			return cmp.Compare(a[1], b[1])
		}
		return result
	})
	return s
}

func (b *Baseware) GetBalloon(balloon string) (ghost.Balloon, error) {
	var ba ghost.Balloon
	var ok bool
	b.balloons.Lock(func(m map[string]ghost.Balloon) {
		ba, ok = m[balloon]
	})
	if !ok {
		return ghost.Balloon{}, errors.New(fmt.Sprintf("balloon not found: %s", balloon))
	}
	return ba, nil
}

func (b *Baseware) GetPreferences(s string) (string, bool) {
	ret := ""
	ok := false
	b.preferences.RLock(func(m map[string]string) {
		ret, ok = m[s]
	})
	return ret, ok
}

func (b *Baseware) SetPreferences(k string, v string) {
	b.preferences.Lock(func(m map[string]string) {
		m[k] = v
	})
}

func (b *Baseware) OpenScriptLog() {
	err := b.scriptlogProcess.Exec()
	if err != nil {
		log.Println(err)
	}
}

func (b *Baseware) AppendScriptLog(_ string, data structs.LogData) {
	s := make([]string, 0, len(data.Status))
	for k, _ := range data.Status {
		s = append(s, k)
	}
	slices.SortFunc(s, func(a string, b string) int {
		return cmp.Compare(a, b)
	})
	b.scriptlogProcess.Append(fmt.Sprintf("%s\x01%s\x01%s\x01%s", data.Time.Format("2006-01-02 15:04:05"), data.Ghost, strings.Join(s, ","), data.Script))
}

func (b *Baseware) OpenPreferences() {
	err := b.preferencesProcess.Exec(b)
	if err != nil {
		log.Println(err)
	}
}

func (b *Baseware) NotifyConfigurationChanged() {
	b.sakuras.RLock(func(m map[string]*sakura.Sakura) {
		for _, v := range m {
			v.OnConfigurationChanged()
		}
	})
}
