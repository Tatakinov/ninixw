package baseware

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"ninixw-core/path"
)

type scriptlog struct {
	mutex sync.RWMutex
	alive bool
	cmd *exec.Cmd
	cancel context.CancelFunc
	ch chan string
	buffer []string
}

func newScriptlog() *scriptlog {
	return &scriptlog{
		mutex: sync.RWMutex{},
		alive: false,
		buffer: make([]string, 0, 50),
	}
}

func (s *scriptlog) Exec() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.alive {
		return nil
	}
	s.ch = make(chan string, 1)
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.cmd = exec.CommandContext(ctx, filepath.Join(path.GuiDir(), "ninixw-scriptlog.exe"))
	w, err := s.cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		defer w.Close()
		for _, v := range s.buffer {
			fmt.Fprintln(w, v)
		}
		for {
			data, ok := <- s.ch
			if !ok {
				break
			}
			fmt.Fprintln(w, data)
		}
	}()
	r, err := s.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	go func() {
		defer s.Stop()
		defer r.Close()
		reader := bufio.NewReader(r)
		reader.ReadLine()
	}()
	err = s.cmd.Start()
	if err != nil {
		return err
	}
	s.alive = true
	return nil
}

func (s *scriptlog) Append(data string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.buffer = append(s.buffer, data)
	if len(s.buffer) > 50 {
		s.buffer = s.buffer[1:]
	}
	if !s.alive {
		return
	}
	s.ch <- data
}

func (s *scriptlog) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if !s.alive {
		return nil
	}
	s.alive = false
	close(s.ch)
	ch := make(chan error)
	go func() {
		ch <- s.cmd.Wait()
	}()
	select {
	case err := <- ch:
		return err
	case <- time.After(1 * time.Second):
		s.cancel()
		err := <- ch
		return err
	}
}
