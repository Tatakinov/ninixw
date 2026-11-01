package ghost

type baseInfo struct {
	Key string
	Path string
	Descript map[string]string
}

func (b baseInfo) Get(key string) string {
	return b.Descript[key]
}
