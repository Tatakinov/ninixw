package path

import (
	"os"
	"path/filepath"
)

func HomeDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(dir, ".ninix")
}

func AoDir() string {
	path := os.Getenv("NINIXW_AO_PATH")
	if path == "" {
		return ExecutableDir()
	}
	return path
}

func AiDir() string {
	path := os.Getenv("NINIXW_AI_PATH")
	if path == "" {
		return ExecutableDir()
	}
	return path
}

func SorakadoDir() string {
	path := os.Getenv("NINIXW_SORAKADO_PATH")
	if path == "" {
		return ExecutableDir()
	}
	return path
}

func ProxyDir() string {
	path := os.Getenv("NINIXW_PROXY_PATH")
	if path == "" {
		return ExecutableDir()
	}
	return path
}

func ShmDir() string {
	path := os.Getenv("NINIXW_SHM_PATH")
	if path == "" {
		return ExecutableDir()
	}
	return path
}

func Lib(lib string, dir string) string {
	return filepath.Base(lib)
}
