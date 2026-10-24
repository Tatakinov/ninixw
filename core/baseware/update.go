package baseware

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"uuid"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"ninixw-core/interfaces"
	"ninixw-core/log"
	"ninixw-core/utils"
)

type updateFile struct {
	Path string
	Hash string
	Size int
	Date int64
}

func (b *Baseware) UpdateGhost(url string, root string, s interfaces.ISakura) {
	log.Println(root)
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
		return
	}
	defer res.Body.Close()
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return
	}
	data := string(bytes)
	info, err := parseUpdateFiles(data)
	if err != nil {
		log.Println(err)
		return
	}
	for _, f := range info {
		log.Println(f.Path)
	}
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
		path := slice[0]
		hash := slice[1]
		size := -1
		date := int64(-1)
		if len(slice) > 2 {
			for _, v := range slice[2:] {
				kv := strings.SplitN(v, "=", 2)
				if len(kv) != 2 {
					continue
				}
				switch kv[0] {
				case "size":
					size, err = strconv.Atoi(kv[1])
					if err != nil {
						log.Println(err)
						size = -1
					}
				case "date":
					time, err := time.Parse("2006-01-02T15:04:05", kv[1])
					if err != nil {
						log.Println(err)
					} else {
						date = time.Unix()
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
