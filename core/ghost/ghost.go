package ghost

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"ninixw-core/path"
)

type baseInfo struct {
	Key string
	Path string
	Descript map[string]string
}

type Shell struct {
	baseInfo
}

type Balloon struct {
	baseInfo
}

type Ghost struct {
	baseInfo
	Shells map[string]Shell
}

func NewGhost(ghost_path string) (Ghost, error) {
	file := filepath.Join(ghost_path, "ghost", "master", "descript.txt")
	desc, err := readDescript(file)
	if err != nil {
		return Ghost{}, err
	}
	shells := map[string]Shell{}
	shell_dirs := path.Dirs(filepath.Join(ghost_path, "shell"))
	for _, dir := range shell_dirs {
		desc, err := readDescript(filepath.Join(ghost_path, "shell", dir, "descript.txt"))
		if err != nil {
			continue
		}
		shells[dir] = Shell{ baseInfo: baseInfo {Key: dir, Path: filepath.Join(ghost_path, "shell", dir), Descript: desc } }
	}
	if len(shells) == 0 {
		return Ghost {}, errors.New("no valid shell")
	}
	return Ghost { baseInfo: baseInfo{ Key: filepath.Base(ghost_path), Path: filepath.Join(ghost_path, "ghost", "master"), Descript: desc }, Shells: shells }, nil
}

func readDescript(p string) (map[string]string, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	reader := transform.NewReader(f, japanese.ShiftJIS.NewDecoder())
	scanner := bufio.NewScanner(reader)
	charset := "Shift_JIS"
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "charset") {
			continue
		}
		a := strings.SplitN(line, ",", 2)
		if len(a) < 2 {
			continue
		}
		charset = strings.TrimSpace(a[1])
		break
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	if (charset == "Shift_JIS") {
		reader := transform.NewReader(f, japanese.ShiftJIS.NewDecoder())
		scanner = bufio.NewScanner(reader)
	} else {
		scanner = bufio.NewScanner(f)
	}
	m := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		a := strings.SplitN(line, ",", 2)
		if len(a) < 2 {
			continue
		}
		m[strings.TrimSpace(a[0])] = strings.TrimSpace(a[1])
	}
	return m, nil
}

func NewBalloon(p string) (Balloon, error) {
	desc, err := readDescript(filepath.Join(p, "descript.txt"))
	if err != nil {
		return Balloon{}, err
	}
	return Balloon{ baseInfo: baseInfo{ Key: filepath.Base(p), Path: p, Descript: desc } }, nil
}
