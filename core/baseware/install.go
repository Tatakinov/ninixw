package baseware

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"uuid"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"ninixw-core/interfaces"
	"ninixw-core/log"
	"ninixw-core/path"
	"ninixw-core/structs"
	"ninixw-core/utils"
)

func (b *Baseware) InstallNar(filename string, s interfaces.ISakura) {
	s.SendEvent(structs.Event{
		Method: "GET",
		Option: make(map[string]struct{}),
		Event: "OnInstallBegin",
		Reference: []string{},
	})
	running := false
	b.installs.Lock(func(m map[string]struct{}) {
		_, ok := m[filename]
		if ok {
			running = true
		} else {
			m[filename] = struct{}{}
		}
	})
	if running {
		s.SendEvent(structs.Event{
			Method: "GET",
			Option: make(map[string]struct{}),
			Event: "OnInstallFailure",
			Reference: []string{ "executing" },
		})
		return
	}
	defer func() {
		b.installs.Lock(func(m map[string]struct{}) {
			delete(m, filename)
		})
	}()
	u := ""
	b.uuids.Lock(func(m map[string]struct{}) {
		for {
			u = uuid.New().String()
			_, ok := m[u]
			if ok {
				continue
			}
			m[u] = struct{}{}
			break
		}
	})
	defer func() {
		b.uuids.Lock(func(m map[string]struct{}) {
			delete(m, u)
		})
	}()

	root := filepath.Join(path.TmpDir(), u)
	err := os.RemoveAll(root)
	if err != nil {
		log.Println(err)
		s.SendEvent(structs.Event{
			Method: "GET",
			Option: make(map[string]struct{}),
			Event: "OnInstallFailure",
			Reference: []string{ "unknown" },
		})
		return
	}
	r, err := zip.OpenReader(filename)
	if err != nil {
		log.Println(err)
		s.SendEvent(structs.Event{
			Method: "GET",
			Option: make(map[string]struct{}),
			Event: "OnInstallFailure",
			Reference: []string{ "extraction" },
		})
		return
	}
	defer r.Close()
	for _, f := range r.File {
		name := strings.ReplaceAll(f.Name, `\`, "/")
		if f.NonUTF8 {
			name, _, err = transform.String(japanese.ShiftJIS.NewDecoder(), name)
			if err != nil {
				log.Println(err)
				s.SendEvent(structs.Event{
					Method: "GET",
					Option: make(map[string]struct{}),
					Event: "OnInstallFailure",
					Reference: []string{ "unknown" },
				})
				return
			}
		}
		if !filepath.IsLocal(name) {
			log.Println(fmt.Sprintf("ninixw: invalid filename: %s", name))
			s.SendEvent(structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnInstallFailure",
				Reference: []string{ "unknown" },
			})
			return
		}
		reader, err := f.Open()
		if err != nil {
			log.Println(err)
			s.SendEvent(structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnInstallFailure",
				Reference: []string{ "unknown" },
			})
			return
		}
		defer reader.Close()
		p := filepath.Join(root, name)
		err = os.MkdirAll(filepath.Dir(p), 0755)
		if err != nil {
			log.Println(err)
			s.SendEvent(structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnInstallFailure",
				Reference: []string{ "unknown" },
			})
			return
		}
		if strings.HasSuffix(name, "/") || strings.HasSuffix(name, `\`) {
			// directory
			continue
		}
		o, err := os.OpenFile(p, os.O_RDWR | os.O_CREATE, 0644)
		if err != nil {
			log.Println(err)
			s.SendEvent(structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnInstallFailure",
				Reference: []string{ "unknown" },
			})
			return
		}
		defer o.Close()
		_, err = io.Copy(o, reader)
		if err != nil {
			s.SendEvent(structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnInstallFailure",
				Reference: []string{ "unknown" },
			})
			log.Println(err)
		}
	}
	// うっかりフォルダごとNARにしてしまった場合は1階層掘る
	_, err = os.Stat(filepath.Join(root, "install.txt"))
	if err != nil {
		dirs := path.Dirs(root)
		if len(dirs) == 1 {
			root = filepath.Join(root, dirs[0])
		}
	}
	id, name, path, err := install(root, s, false)
	if err != nil {
		s.SendEvent(structs.Event{
			Method: "GET",
			Option: make(map[string]struct{}),
			Event: "OnInstallFailure",
			Reference: []string{ "unknown" },
		})
	} else {
		s.SendEvent(structs.Event{
			Method: "GET",
			Option: make(map[string]struct{}),
			Event: "OnInstallCompleteEX",
			Reference: []string{ id, name, path },
		}, structs.Event{
			Method: "GET",
			Option: make(map[string]struct{}),
			Event: "OnInstallComplete",
			Reference: []string{
				strings.Split(id, "\x01")[0],
				strings.Split(name, "\x01")[0],
			},
		})
	}
}

func install(root string, s interfaces.ISakura, in bool) (string, string, string, error) {
	m, err := utils.ReadSetting(filepath.Join(root, "install.txt"))
	if err != nil {
		log.Println(err)
		s.SendEvent(structs.Event{
			Method: "GET",
			Option: make(map[string]struct{}),
			Event: "OnInstallFailure",
			Reference: []string{ "unknown" },
		})
		return "", "", "", err
	}
	switch m["type"] {
	case "ghost":
		b_dir := m["balloon.directory"]
		b_src_dir := m["balloon.source.directory"]
		with_balloon := b_dir != "" && b_src_dir != ""
		if with_balloon {
			dst := filepath.Join(path.BalloonDir(), b_dir)
			if confirm(dst) {
				src := filepath.Join(root, b_src_dir)
				utils.CopyAll(dst, src)
				os.RemoveAll(src)
			} else {
				// TODO cancel
				return "", "", "", errors.New("artificial")
			}
		}
		dir := m["directory"]
		dst := filepath.Join(path.GhostDir(), dir)
		if confirm(dst) {
			err := utils.CopyAll(dst, root)
			if err != nil {
				log.Println(err)
				return "", "", "", err
			}
			if in {
				break
			}
			if with_balloon {
				// TODO stub
			} else {
				s.SendEvent(structs.Event{
					Method: "GET",
					Option: make(map[string]struct{}),
					Event: "OnInstallCompleteEx",
					Reference: []string{ "ghost", m["name"], dst },
				}, structs.Event{
					Method: "GET",
					Option: make(map[string]struct{}),
					Event: "OnInstallComplete",
					Reference: []string{ "ghost", m["name"] },
				})
				return "ghost", m["name"], dst, nil
			}
		} else {
			s.SendEvent(structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnInstallFailure",
				Reference: []string{ "artificial" },
			})
			return "", "", "", errors.New("cancel installation")
		}
	case "shell":
		// TODO stub
	case "supplement":
		// TODO stub
	case "package":
		if in {
			err = errors.New("package in package")
			log.Println(err)
			return "", "", "", err
		}
		// TODO stub
	}
	err = errors.New("invalid type")
	log.Println(err)
	return "", "", "", err
}

func confirm(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return true
	}
	// TODO yes-no
	return false
}
