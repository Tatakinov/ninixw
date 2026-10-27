package baseware

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"ninixw-core/interfaces"
	"ninixw-core/log"
	"ninixw-core/path"
	"ninixw-core/utils"
)

type preferencesData struct {
	FontName string `json:"font_name"`
	FontSize string `json:"font_size"`
	ShellScale string `json:"shell_scale"`
	DefaultBalloon string `json:"default_balloon"`
	Balloons []string `json:"balloons"`
}

func newPreferencesData() preferencesData {
	p := preferencesData{
		FontName: "",
		FontSize: "12",
		ShellScale: "100",
		DefaultBalloon: "",
		Balloons: make([]string, 0),
	}
	return p
}

type preferences struct {
	mutex sync.RWMutex
	alive bool
	cmd *exec.Cmd
	cancel context.CancelFunc
}

func newPreferences() *preferences {
	return &preferences{
		mutex: sync.RWMutex{},
		alive: false,
	}
}

func (p *preferences) Exec(b interfaces.IBaseware) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.alive {
		return nil
	}
	data := newPreferencesData()
	v, ok := b.GetPreferences("font_name")
	if ok {
		data.FontName = v
	}
	v, ok = b.GetPreferences("font_size")
	if ok {
		n, err := strconv.Atoi(v)
		if err == nil {
			data.FontSize = strconv.Itoa(n)
		}
	}
	v, ok = b.GetPreferences("shell_scale")
	if ok {
		n, err := strconv.Atoi(v)
		if err == nil {
			data.ShellScale = strconv.Itoa(n)
		}
	}
	v, ok = b.GetPreferences("default_balloon")
	if ok {
		data.DefaultBalloon = v
	}
	balloons := b.GetBalloons()
	data.Balloons = make([]string, 0, len(balloons))
	for _, v := range balloons {
		data.Balloons = append(data.Balloons, v[0])
	}
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	p.cmd = exec.CommandContext(ctx, filepath.Join(path.GuiDir(), "ninixw-preferences.exe"))
	w, err := p.cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		defer w.Close()
		bytes, err := json.Marshal(data)
		if err != nil {
			log.Println(err)
			return
		}
		w.Write(bytes)
	}()
	r, err := p.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	go func() {
		defer p.Stop()
		defer r.Close()
		reader := bufio.NewReader(r)
		for {
			bytes, err := utils.ReadLine(reader)
			if err != nil {
				if err != io.EOF {
					log.Println(err)
				}
				break
			}
			err = json.Unmarshal(bytes, &data)
			if err != nil {
				log.Println(err)
				continue
			}
			if data.FontName != "" {
				b.SetPreferences("font_name", data.FontName)
			}
			b.SetPreferences("font_size", data.FontSize)
			b.SetPreferences("shell_scale", data.ShellScale)
			if data.DefaultBalloon != "" {
				b.SetPreferences("default_balloon", data.DefaultBalloon)
			}
			b.NotifyConfigurationChanged()
		}
	}()
	err = p.cmd.Start()
	if err != nil {
		return err
	}
	p.alive = true
	return nil
}

func (p *preferences) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if !p.alive {
		return nil
	}
	p.alive = false
	ch := make(chan error)
	go func() {
		ch <- p.cmd.Wait()
	}()
	select {
	case err := <- ch:
		return err
	case <- time.After(1 * time.Second):
		p.cancel()
		err := <- ch
		return err
	}
}
