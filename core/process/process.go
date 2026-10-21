package process

import (
	"context"
	"encoding/binary"
	"io"
	"os/exec"
	"time"
)

type Data struct {
	Required bool
	Channel chan <- string
	Content string
}

type Process struct {
	alive chan struct{}
	cmd *exec.Cmd
	send chan Data
	Recv chan string
	cancel context.CancelFunc
}

func NewProcess(name string, args ...string) (Process, error) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, name, args...)
	queue := make(chan Data, 100)
	send := make(chan Data, 100)
	w, err := cmd.StdinPipe()
	if err != nil {
		return Process{}, err
	}
	go func() {
		defer w.Close()
		for {
			data, ok := <- send
			if ok && len(data.Content) > 0 {
				queue <- data
				var length uint32 = uint32(len(data.Content))
				buf := make([]byte, 4)
				binary.LittleEndian.PutUint32(buf, length)
				w.Write(buf)
				buf = []byte(data.Content)
				w.Write(buf)
			} else {
				close(queue)
				var length uint32 = 0
				buf := make([]byte, 4)
				binary.LittleEndian.PutUint32(buf, length)
				w.Write(buf)
				break
			}
		}
	}()
	r, err := cmd.StdoutPipe()
	if err != nil {
		return Process{}, err
	}
	go func() {
		defer r.Close()
		for {
			recv, ok := <- queue
			if !ok {
				return
			}
			defer close(recv.Channel)
			buf := make([]byte, 4)
			n, err := r.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				return
			} else if n == 4 {
				length := binary.LittleEndian.Uint32(buf)
				buf := make([]byte, length)
				total := 0
				for {
					n, err := r.Read(buf[total:])
					total += n
					if err != nil {
						return
					} else if uint32(total) < length {
						continue
					} else if uint32(total) == length {
						data := string(buf)
						if recv.Required {
							recv.Channel <- data
						}
						break
					} else {
						return
					}
				}
			} else {
				return
			}
		}
	}()
	err = cmd.Start()
	if err != nil {
		return Process{}, err
	}
	return Process{
		alive: make(chan struct{}),
		cmd: cmd,
		send: send,
		cancel: cancel,
	}, nil
}

func (p Process) Send(required bool, channel chan <- string, content string) {
	select {
	case <- p.alive:
		return
	default:
		// nop
	}
	p.send <- Data{ Required: required, Channel: channel, Content: content }
}

func (p Process) Stop() error {
	close(p.alive)
	close(p.send)
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
