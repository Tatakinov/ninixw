package main

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"ninixw-core/ghost"
	"ninixw-core/path"
	"ninixw-core/with_mutex"
)

type Baseware struct {
	ghosts with_mutex.WithMutex[string, ghost.Ghost]
	balloons with_mutex.WithMutex[string, ghost.Balloon]
	sakuras with_mutex.WithMutex[string, *ghost.Sakura]
	wg sync.WaitGroup
	channel chan string
	uuids with_mutex.WithMutex[string, struct{}]
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
		sakuras: with_mutex.NewWithMutex[string, *ghost.Sakura](),
		wg: sync.WaitGroup{},
		channel: make(chan string),
		uuids: with_mutex.NewWithMutex[string, struct{}](),
	}
}

func (b *Baseware) Load() {
	dirs := path.Dirs(path.GhostDir())
	for _, dir := range dirs {
		g, err := ghost.NewGhost(filepath.Join(path.GhostDir(), dir))
		if err != nil {
			continue
		}
		b.ghosts.Lock(func(m map[string]ghost.Ghost) {
			m[dir] = g
		})
	}
	dirs = path.Dirs(path.BalloonDir())
	for _, dir := range dirs {
		balloon, err := ghost.NewBalloon(filepath.Join(path.BalloonDir(), dir))
		if err != nil {
			continue
		}
		b.balloons.Lock(func(m map[string]ghost.Balloon) {
			m[dir] = balloon
		})
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
			ghost_name = "knowledgeable_kernel"
		}
		err := b.StartSakura(ghost_name)
		if err != nil {
			fmt.Println(err)
		}
	}
	b.wg.Wait()
}

func (b *Baseware) StartSakura(name string) error {
	var g ghost.Ghost
	var balloon ghost.Balloon
	var ok bool
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
	sakura, err := ghost.NewSakura(b, g, s, balloon)
	if err != nil {
		return err
	}
	b.sakuras.Lock(func(m map[string]*ghost.Sakura) {
		m["knowledgeable_kernel"] = sakura
	})
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		sakura.MainLoop()
	}()
	return nil
}

func (b *Baseware) StopSakura(name string) {
	b.sakuras.Lock(func(m map[string]*ghost.Sakura) {
		delete(m, name)
	})
}

func (b *Baseware) RegisterUuid(uuid string) bool {
	ret := false
	b.uuids.Lock(func(m map[string]struct{}) {
		_, ok := m[uuid]
		if ok {
			return
		}
		m[uuid] = struct{}{}
		ret = true
	})
	return ret
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
	b.sakuras.Lock(func(m map[string]*ghost.Sakura) {
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
