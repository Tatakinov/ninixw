package path

import (
	"os"
	"path/filepath"
	"strings"
)

func AppendDirSeparator(path string) string {
	if path == "" {
		panic("null path")
	}
	sep := string(filepath.Separator)
	if strings.HasSuffix(path, sep) {
		return path
	}
	return path + sep
}

func ExecutableDir() string {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	dir := filepath.Dir(exe)
	return dir
}

func GhostDir() string {
	return filepath.Join(HomeDir(), "ghost")
}

func BalloonDir() string {
	return filepath.Join(HomeDir(), "balloon")
}

func PluginDir() string {
	return filepath.Join(HomeDir(), "plugin")
}

func SockDir() string {
	return filepath.Join(HomeDir(), "sock")
}

func TmpDir() string {
	return filepath.Join(HomeDir(), "tmp")
}

func Dirs(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	dirs := []string{}
	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := os.Stat(filepath.Join(dir, entry.Name()))
			if err != nil || !info.IsDir() {
				continue
			}
		}
		dirs = append(dirs, entry.Name())
	}
	return dirs
}
