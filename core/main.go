package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"ninixw-core/baseware"
	"ninixw-core/log"
	"ninixw-core/path"
)

func main() {
	ghost_name := flag.String("ghost", "", "ghost name")
	flag.Parse()
	// shm
	shm_exe := filepath.Join(path.ShmDir(), "ninixw-shm.exe")
	shm_bytes, err := exec.Command(shm_exe, "-g").Output()
	create_new := false
	if err != nil {
		log.Println(err)
		create_new = true
	}
	shm_output := strings.TrimRight(string(shm_bytes), "\r\n")
	if shm_output != "" {
		conn, err := net.Dial("unix", shm_output)
		if err != nil {
			create_new = true
		} else {
			defer conn.Close()
			if *ghost_name != "" {
				w := bufio.NewWriter(conn)
				fmt.Fprintln(w, fmt.Sprintf("CallGhost:%s", *ghost_name))
				w.Flush()
			}
			if *ghost_name == "" {
				fmt.Println("ninixw is already running")
			}
			return
		}
	}
	args := make([]string, 1)
	if create_new {
		args = append(args, "-n")
	}
	args = append(args, filepath.Join(path.SockDir(), "ninix"))
	ctx, cancel := context.WithCancel(context.Background())
	shm_cmd := exec.CommandContext(ctx, shm_exe, args...)
	shm_write, err := shm_cmd.StdinPipe()
	if err != nil {
		log.Println(err)
		return
	}
	shm_cmd.Start()

	baseware := baseware.NewBaseware()
	baseware.Load()
	baseware.Run(*ghost_name)

	shm_write.Close()
	shm_ch := make(chan error)
	go func() {
		shm_ch <- shm_cmd.Wait()
	}()
	select {
	case <- shm_ch:
		return
	case <- time.After(1 * time.Second):
		cancel()
		<- shm_ch
		return
	}
}
