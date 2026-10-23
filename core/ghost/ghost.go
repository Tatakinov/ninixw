package ghost

import (
	"errors"
	"path/filepath"

	"ninixw-core/path"
	"ninixw-core/utils"
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
	desc, err := utils.ReadSetting(file)
	if err != nil {
		return Ghost{}, err
	}
	shells := map[string]Shell{}
	shell_dirs := path.Dirs(filepath.Join(ghost_path, "shell"))
	for _, dir := range shell_dirs {
		desc, err := utils.ReadSetting(filepath.Join(ghost_path, "shell", dir, "descript.txt"))
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

func NewBalloon(p string) (Balloon, error) {
	desc, err := utils.ReadSetting(filepath.Join(p, "descript.txt"))
	if err != nil {
		return Balloon{}, err
	}
	return Balloon{ baseInfo: baseInfo{ Key: filepath.Base(p), Path: p, Descript: desc } }, nil
}
