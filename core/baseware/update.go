package baseware

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"uuid"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"ninixw-core/interfaces"
	"ninixw-core/log"
	"ninixw-core/path"
	"ninixw-core/structs"
	"ninixw-core/utils"
)

type updateFile struct {
	Path string
	Hash string
	Size int64
	Date int64
}

func (b *Baseware) UpdateGhost(url string, root string, s interfaces.ISakura, reason string) {
	running := false
	b.installs.Lock(func(m map[string]struct{}) {
		_, ok := m[s.GhostPath()]
		if ok {
			running = true
		} else {
			m[s.GhostPath()] = struct{}{}
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
			delete(m, s.GhostPath())
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
	tmp_dir := filepath.Join(path.TmpDir(), u)
	err := os.RemoveAll(tmp_dir)

	s.SendEvent(structs.Event{
		Method: "GET",
		Option: make(map[string]struct{}),
		Event: "OnUpdateBegin",
		Reference: []string{
			s.GhostName(),
			root,
			"",
			"ghost",
			reason,
		},
	})

	if !strings.HasSuffix(url, "/") {
		url = fmt.Sprintf("%s/", url)
	}
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	res, err := client.Get(fmt.Sprintf("%s%s", url, "updates2.dau"))
	if err != nil {
		log.Println(err)
		s.SendEvent(structs.Event{
			Method: "GET",
			Option: make(map[string]struct{}),
			Event: "OnUpdateFailure",
			Reference: []string{
				"unknown",
			},
		})
		return
	}
	if res.StatusCode != 200 {
		log.Println(fmt.Sprintf("%s: updates2.dau", res.Status))
		s.SendEvent(structs.Event{
			Method: "GET",
			Option: make(map[string]struct{}),
			Event: "OnUpdateFailure",
			Reference: []string{
				strconv.Itoa(res.StatusCode),
			},
		})
	}
	defer res.Body.Close()
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		s.SendEvent(structs.Event{
			Method: "GET",
			Option: make(map[string]struct{}),
			Event: "OnUpdateFailure",
			Reference: []string{
				"unknown",
			},
		})
		return
	}
	data := string(bytes)
	info, err := parseUpdateFiles(data)
	if err != nil {
		log.Println(err)
		s.SendEvent(structs.Event{
			Method: "GET",
			Option: make(map[string]struct{}),
			Event: "OnUpdateFailure",
			Reference: []string{
				"unknown",
			},
		})
		return
	}
	info, err = updateRequired(info, root)
	if err != nil {
		log.Println(err)
		s.SendEvent(structs.Event{
			Method: "GET",
			Option: make(map[string]struct{}),
			Event: "OnUpdateFailure",
			Reference: []string{
				"unknown",
			},
		})
		return
	}
	if len(info) == 0 {
		s.SendEvent(structs.Event{
			Method: "GET",
			Option: make(map[string]struct{}),
			Event: "OnUpdateComplete",
			Reference: []string{
				"none",
				"",
				"",
				"ghost",
				reason,
			},
		})
		return
	}
	list := make([]string, 0, len(info))
	for _, f := range info {
		list = append(list, f.Path)
	}
	s.SendEvent(structs.Event{
		Method: "GET",
		Option: make(map[string]struct{}),
		Event: "OnUpdateReady",
		Reference: []string{
			strconv.Itoa(len(list) - 1),
			strings.Join(list, ","),
			"",
			"ghost",
			reason,
		},
	})
	for i, f := range info {
		s.SendEvent(structs.Event{
			Method: "GET",
			Option: make(map[string]struct{}),
			Event: "OnUpdate.OnDownloadBegin",
			Reference: []string{
				f.Path,
				strconv.Itoa(i),
				strconv.Itoa(len(info) - 1),
				"ghost",
				reason,
			},
		})
		res, err := client.Get(fmt.Sprintf("%s%s", url, f.Path))
		if err != nil {
			log.Println(err)
			s.SendEvent(structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnUpdateFailure",
				Reference: []string{
					"unknown",
				},
			})
			return
		}
		defer res.Body.Close()
		p := filepath.Join(tmp_dir, f.Path)
		err = os.MkdirAll(filepath.Dir(p), 0755)
		if err != nil {
			log.Println(err)
			s.SendEvent(structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnUpdateFailure",
				Reference: []string{
					"unknown",
				},
			})
			return
		}
		o, err := os.OpenFile(p, os.O_RDWR | os.O_CREATE, 0644)
		if err != nil {
			log.Println(err)
			s.SendEvent(structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnUpdateFailure",
				Reference: []string{
					"unknown",
				},
			})
			return
		}
		defer o.Close()
		_, err = io.Copy(o, res.Body)
		if err != nil {
			log.Println(err)
			s.SendEvent(structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnUpdateFailure",
				Reference: []string{
					"unknown",
				},
			})
			return
		}
		bytes, err := os.ReadFile(p)
		if err != nil {
			log.Println(err)
			s.SendEvent(structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnUpdateFailure",
				Reference: []string{
					"unknown",
				},
			})
			return
		}
		hb := md5.Sum(bytes)
		h := hex.EncodeToString(hb[:])
		s.SendEvent(structs.Event{
			Method: "GET",
			Option: make(map[string]struct{}),
			Event: "OnUpdate.OnMD5CompareBegin",
			Reference: []string{
				f.Path,
				f.Hash,
				h,
				"ghost",
				reason,
			},
		})
		if h == f.Hash {
			s.SendEvent(structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnUpdate.OnMD5CompareComplete",
				Reference: []string{
					f.Path,
					f.Hash,
					h,
					"ghost",
					reason,
				},
			})
		} else {
			s.SendEvent(structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnUpdate.OnMD5CompareFailure",
				Reference: []string{
					f.Path,
					f.Hash,
					h,
					"ghost",
					reason,
				},
			})
			s.SendEvent(structs.Event{
				Method: "GET",
				Option: make(map[string]struct{}),
				Event: "OnUpdateFailure",
				Reference: []string{
					"md5 miss",
				},
			})
			return
		}
	}
	err = utils.CopyAll(root, tmp_dir)
	if err != nil {
		log.Println(err)
		s.SendEvent(structs.Event{
			Method: "GET",
			Option: make(map[string]struct{}),
			Event: "OnUpdateFailure",
			Reference: []string{
				"unknown",
			},
		})
		return
	}
	s.ReloadShiori()
	s.ReloadShell()
	s.ReloadBalloon()
	s.SendEvent(structs.Event{
		Method: "GET",
		Option: make(map[string]struct{}),
		Event: "OnUpdateComplete",
		Reference: []string{
			"changed",
			strings.Join(list, ","),
			"ghost",
			reason,
		},
	})
}

func parseUpdateFiles(data string) ([]updateFile, error) {
	info := make([]updateFile, 0)
	r := bufio.NewReader(strings.NewReader(data))
	bytes, err := utils.ReadLine(r)
	if err != nil {
		return []updateFile{}, err
	}
	line := string(bytes)
	slice := strings.Split(line, "\x01")
	if len(slice) > 2 {
		charset := "UTF-8"
		for _, v := range slice[2:] {
			kv := strings.SplitN(v, "=", 2)
			if len(kv) != 2 {
				continue
			}
			if kv[0] == "charset" {
				charset = kv[1]
				break
			}
		}
		switch charset {
		case "UTF-8", "ASCII":
			r = bufio.NewReader(strings.NewReader(data))
		case "Shift_JIS":
			r = bufio.NewReader(transform.NewReader(strings.NewReader(data), japanese.ShiftJIS.NewDecoder()))
		}
	} else {
		r = bufio.NewReader(strings.NewReader(data))
	}
	for {
		bytes, err := utils.ReadLine(r)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return []updateFile{}, err
			}
		}
		line := string(bytes)
		slice := strings.Split(line, "\x01")
		if len(slice) < 2 {
			continue
		}
		path := strings.ReplaceAll(slice[0], `\`, "/")
		hash := slice[1]
		size := int64(-1)
		date := int64(-1)
		if len(slice) > 2 {
			for _, v := range slice[2:] {
				kv := strings.SplitN(v, "=", 2)
				if len(kv) != 2 {
					continue
				}
				switch kv[0] {
				case "size":
					tmp, err := strconv.Atoi(kv[1])
					if err != nil {
						log.Println(err)
					} else {
						size = int64(tmp)
					}
				case "date":
					t, err := time.Parse("2006-01-02T15:04:05", kv[1])
					if err != nil {
						log.Println(err)
					} else {
						date = t.Unix()
					}
				}
			}
		}
		if !filepath.IsLocal(path) {
			return []updateFile{}, errors.New("invalid path")
		}
		info = append(info, updateFile{Path: path, Hash: hash, Size: size, Date: date})
	}
	return info, nil
}

func updateRequired(files []updateFile, root string) ([]updateFile, error) {
	ret := make([]updateFile, 0)
	for _, f := range files {
		p := filepath.Join(root, f.Path)
		s, err := os.Stat(p)
		if err != nil {
			ret = append(ret, f)
			continue
		}
		if f.Size != -1 && s.Size() != f.Size {
			ret = append(ret, f)
			continue
		}
		if f.Date != -1 && s.ModTime().Unix() < f.Date {
			ret = append(ret, f)
			continue
		}
		bytes, err := os.ReadFile(p)
		if err != nil {
			return []updateFile{}, err
		}
		hb := md5.Sum(bytes)
		h := hex.EncodeToString(hb[:])
		if f.Hash != h {
			ret = append(ret, f)
			continue
		}
	}
	return ret, nil
}
