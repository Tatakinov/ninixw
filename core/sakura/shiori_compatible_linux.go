package sakura

import (
	"fmt"
	"path/filepath"
	"strings"
)

func getShioriFfiInfo(lib string) []string {
	base := filepath.Base(lib)
	switch base {
	case "satori.dll":
		return []string{
			"int load:satori_load(void*, long)",
			"void* request:satori_request(int, void*, void*)",
			"int unload:satori_unload(int)",
		}
	case "yaya.dll":
		return []string{
			"long load:multi_load(void*, long)",
			"void* request:multi_request(long, void*, void*)",
			"int unload:multi_unload(long)",
		}
	default:
		stem := strings.TrimSuffix(base, filepath.Ext(base))
		return []string{
			fmt.Sprintf("long load:%s_load(void*, long)", stem),
			fmt.Sprintf("void* request:%s_request(long, void*, void*)", stem),
			fmt.Sprintf("int unload:%s_unload(long)", stem),
		}
	}
}
