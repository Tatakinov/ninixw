package path

import (
	"os"
	"path/filepath"
)

func libexecDir() string {
	return filepath.Join(ExecutableDir(), "../libexec")
}

func HomeDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(dir, ".ninix")
}

func AoDir() string {
	path := os.Getenv("NINIXW_LIBEXEC_PATH")
	if path == "" {
		return libexecDir()
	}
	return path
}

func AiDir() string {
	path := os.Getenv("NINIXW_LIBEXEC_PATH")
	if path == "" {
		return libexecDir()
	}
	return path
}

func SorakadoDir() string {
	path := os.Getenv("NINIXW_LIBEXEC_PATH")
	if path == "" {
		return libexecDir()
	}
	return path
}

func SorakadoStem(_ string) string {
	return ""
}

func ProxyDir() string {
	path := os.Getenv("NINIXW_LIBEXEC_PATH")
	if path == "" {
		return libexecDir()
	}
	return path
}

func ShmDir() string {
	path := os.Getenv("NINIXW_LIBEXEC_PATH")
	if path == "" {
		return libexecDir()
	}
	return path
}

func GuiDir() string {
	path := os.Getenv("NINIXW_LIBEXEC_PATH")
	if path == "" {
		return libexecDir()
	}
	return path
}

func Lib(lib string, dir string) string {
	return filepath.Base(lib)
}
