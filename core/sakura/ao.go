package sakura

import (
	"errors"
	"path/filepath"

	"ninixw-core/ghost"
	"ninixw-core/path"
	"ninixw-core/process"
)

type ao struct {
	ghost.Shell
	sorakado
}

func newAo(shell ghost.Shell) (ao, error) {
	exe := shell.Descript["ao"]
	if exe == "" {
		exe = "sorakado_builtin.exe"
	}
	s := []string{
		path.AoDir(),
		path.SorakadoDir(),
		path.ExecutableDir(),
	}
	for _, p := range s {
		process, err := process.NewProcess(filepath.Join(p, exe))
		if err != nil {
			continue
		}
		return ao {
			Shell: shell,
			sorakado: sorakado {
				process: process,
			},
		}, nil
	}
	return ao{}, errors.New("sorakado not found")
}

func (ao ao) Stop() error {
	return ao.sorakado.Stop()
}

