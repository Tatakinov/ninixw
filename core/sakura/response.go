package sakura

import (
	"bufio"
	"errors"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"ninixw-core/log"
	"ninixw-core/utils"
)

type Response struct {
	Protocol string
	Code int
	Message string
	Header map[string]string
}

func NewResponse(data string, charset string) (Response, error) {
	iss := strings.NewReader(data)
	reader := bufio.NewReader(iss)
	if charset == "Shift_JIS" {
		reader = bufio.NewReader(transform.NewReader(iss, japanese.ShiftJIS.NewDecoder()))
	}
	bytes, err := utils.ReadLine(reader)
	if err != nil {
		return Response{}, err
	}
	line := string(bytes)
	line = strings.TrimRight(line, "\r\n")
	slice := strings.SplitN(line, " ", 3)
	if len(slice) < 3 {
		return Response{}, errors.New("invalid shiori status line")
	}
	p := slice[0]
	c, err := strconv.Atoi(slice[1])
	if err != nil {
		return Response{}, err
	}
	m := slice[2]
	h := make(map[string]string)
	for {
		bytes, err = utils.ReadLine(reader)
		if err != nil {
			log.Println(err)
			return Response{}, err
		}
		line = string(bytes)
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		slice = strings.SplitN(line, ":", 2)
		if len(slice) != 2 {
			return Response{}, errors.New("invalid shiori header")
		}
		key := strings.TrimSpace(slice[0])
		value := strings.TrimSpace(slice[1])
		h[key] = value
	}
	return Response {
		Protocol: p,
		Code: c,
		Message: m,
		Header: h,
	}, nil
}
