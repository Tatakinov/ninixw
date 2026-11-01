package ghost

import (
	"path/filepath"

	"ninixw-core/utils"
)

type Balloon struct {
	baseInfo
}

func NewBalloon(p string) (Balloon, error) {
	desc, err := utils.ReadSetting(filepath.Join(p, "descript.txt"))
	if err != nil {
		return Balloon{}, err
	}
	return Balloon{ baseInfo: baseInfo{ Key: filepath.Base(p), Path: p, Descript: desc } }, nil
}
