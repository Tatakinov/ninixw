package path

import (
	"path/filepath"
	"strings"
)

func HomeDir() string {
	return filepath.Join(ExecutableDir(), ".ninix")
}

func AoDir() string {
	return filepath.Join(ExecutableDir(), "internal", "ao")
}

func AiDir() string {
	return filepath.Join(ExecutableDir(), "internal", "ai")
}

func SorakadoDir() string {
	return filepath.Join(ExecutableDir(), "internal", "sorakado")
}

func SorakadoStem(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	return stem
}

func ProxyDir() string {
	return filepath.Join(ExecutableDir(), "internal", "proxy")
}

func ShmDir() string {
	return filepath.Join(ExecutableDir(), "internal", "shm")
}

func GuiDir() string {
	return filepath.Join(ExecutableDir(), "internal", "gui")
}

func Lib(lib string, dir string) string {
	return filepath.Join(dir, lib)
}
