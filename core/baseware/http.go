package baseware

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"ninixw-core/log"
	"ninixw-core/structs"
	"ninixw-core/utils"
)

func (b *Baseware) HttpRequest(method string, u string, body string, opt structs.HttpOption, root string) (chan structs.HttpResponse, context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(context.Background())
	req, err := http.NewRequestWithContext(ctx, method, u, nil)
	if body != "" {
		req, err = http.NewRequestWithContext(ctx, method, u, strings.NewReader(body))
	}
	if err != nil {
		log.Println(err)
		return make(chan structs.HttpResponse), cancel, err
	}
	ur, err := url.Parse(u)
	if err != nil {
		log.Println(err)
		return make(chan structs.HttpResponse), cancel, err
	}
	if opt.NoFile == "" && opt.File == "" {
		opt.File = filepath.Base(filepath.Clean(ur.Path))
		if opt.File == "" {
			opt.File = "index.html"
		}
	}
	ch := make(chan structs.HttpResponse, 1)
	go func() {
		defer close(ch)
		res, err := b.httpClient.Do(req)
		if err != nil {
			log.Println(err)
			return
		}
		defer res.Body.Close()
		h := make([]string, 0)
		for k, v := range res.Header {
			h = append(h, fmt.Sprintf("%s: %s", k, strings.Join(v, ",")))
		}
		r := structs.HttpResponse{
			Method: method,
			Url: u,
			Data: "",
			Status: strconv.Itoa(res.StatusCode),
			Cookie: "", // TODO stub
			Headers: strings.Join(h, "\x01"),
		}
		if opt.NoFile != "" {
			buffer := make([]string, 0)
			for {
				// TODO transform.NewReader charset
				bytes, err := utils.ReadLine(bufio.NewReader(res.Body))
				line := string(bytes)
				if err != nil {
					if err != io.EOF {
						log.Println(err)
						return
					}
					if len(bytes) > 0 {
						buffer = append(buffer, line)
					}
					break
				}
				buffer = append(buffer, line)
			}
			r.Data = strings.Join(buffer, "\x01")
		} else {
			filename := filepath.Join(root, "var", opt.File)
			o, err := os.OpenFile(filename, os.O_RDWR | os.O_CREATE | os.O_TRUNC, 0644)
			if err != nil {
				log.Println(err)
				return
			}
			io.Copy(o, res.Body)
			r.Data = filename
		}
		ch <- r
	}()
	return ch, cancel, nil
}
