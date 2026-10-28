package sakura

import (
	"errors"
	"path/filepath"

	"ninixw-core/ghost"
	"ninixw-core/path"
	"ninixw-core/process"
)

type ai struct {
	ghost.Balloon
	sorakado
}

func newAi(balloon ghost.Balloon) (ai, error) {
	exe := balloon.Descript["ai"]
	if exe == "" {
		exe = "sorakado_builtin.exe"
	}
	dir := path.SorakadoStem(exe)
	s := []string{
		path.AiDir(),
		path.SorakadoDir(),
		path.ExecutableDir(),
	}
	for _, p := range s {
		process, err := process.NewProcess(filepath.Join(p, dir, exe))
		if err != nil {
			continue
		}
		return ai {
			Balloon: balloon,
			sorakado: sorakado {
				process: process,
			},
		}, nil
	}
	return ai{}, errors.New("sorakado not found")
}

func (ai ai) Stop() error {
	return ai.sorakado.Stop()
}

