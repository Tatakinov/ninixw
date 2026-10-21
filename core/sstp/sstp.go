package sstp

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type ISSTP interface {
	ProcessSSTP(req IRequest) Response
}

type Proto struct {
	Name string
	Version string
}

func (p Proto) String() string {
	return p.Name + "/" + p.Version
}

type IRequest interface {
	Protocol() Proto
	Sstp() (SstpRequest, error)
	KeepAlive() bool
}

type HttpRequest struct {
	Method string
	Path string
	Proto Proto
	Header map[string]string
	Content string
}

type SstpRequest struct {
	Method string
	Proto Proto
	Header map[string]string
	Script map[string]string
	Reference []string
	Content string
}

type Response struct {
	Code int
	Message string
	Proto Proto
	Header map[string]string
	Content string
}

func (r Response) String() string {
	s := make([]string, 0)
	s = append(s, fmt.Sprintf("%s %d %s", r.Proto.String(), r.Code, r.Message))
	charset, ok := r.Header["Charset"]
	if ok {
		s = append(s, fmt.Sprintf("Charset: %s", charset))
	} else {
		s = append(s, "Charset: UTF-8")
	}
	for k, v := range r.Header {
		if k == "Charset" {
			continue
		}
		s = append(s, fmt.Sprintf("%s: %s", k, v))
	}
	s = append(s, "\r\n")
	return strings.Join(s, "\r\n")
}

func NewSSTPHeader(keep_alive bool) map[string]string {
	conn := "close"
	if keep_alive {
		conn = "keep-alive"
	}
	return map[string]string{
		"Charset": "UTF-8",
		"Connection": conn,
	}
}

func NewResponse(code int, message string, header map[string]string, content string, protocol Proto) Response {
	return Response{
		Code: code,
		Message: message,
		Proto: protocol,
		Header: header,
		Content: content,
	}
}

func HandleSSTP(is ISSTP, server net.Listener) {
	for {
		conn, err := server.Accept()
		if err != nil {
			break
		}
		go func() {
			reader := bufio.NewReader(conn)
			writer := bufio.NewWriter(conn)
			for {
				bytes, _, err := reader.ReadLine()
				if err != nil {
					break
				}
				line := string(bytes)
				req, err := Parse(reader, line)
				if err != nil {
					break
				}
				res := is.ProcessSSTP(req)
				writer.WriteString(res.String())
				writer.Flush()
			}
		}()
	}
}

func parseProtocol(s []string) (string, string, Proto, bool) {
	if len(s) == 2 && strings.HasPrefix(s[1], "SSTP/") {
		s2 := strings.SplitN(s[1], "/", 2)
		return s[0], "", Proto{ Name: s2[0], Version: s2[1] }, true
	}
	if len(s) == 3 && strings.HasPrefix(s[2], "HTTP/") {
		s2 := strings.SplitN(s[2], "/", 2)
		return s[0], s[1], Proto{ Name: s2[0], Version: s2[1] }, true
	}
	return "", "", Proto{}, false
}

func Parse(r *bufio.Reader, line string) (IRequest, error) {
	line = strings.TrimRight(line, "\r\n")
	s := strings.Split(line, " ")
	method, path, protocol, ok := parseProtocol(s)
	if !ok {
		return SstpRequest{}, errors.New("invalid status line")
	}
	switch protocol.Name {
	case "HTTP":
		return parseHTTP(r, method, path, protocol)
	case "SSTP":
		return parseSSTP(r, method, protocol)
	default:
		return SstpRequest{}, errors.New("unknown protocol")
	}
}

func parseHTTP(r *bufio.Reader, method string, path string, protocol Proto) (HttpRequest, error) {
	header := make(map[string]string)
	content := ""
	for {
		bytes, _, err := r.ReadLine()
		if err != nil {
			return HttpRequest{}, err
		}
		line := string(bytes)
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		s := strings.SplitN(line, ":", 2)
		if (len(s) < 2) {
			continue
		}
		k := strings.TrimSpace(s[0])
		v := strings.TrimSpace(s[1])
		header[k] = v
	}
	content_length, ok := header["Content-Length"]
	if ok {
		l, err := strconv.Atoi(content_length)
		if err != nil {
			return HttpRequest{}, err
		}
		if l > 0 {
			bytes := make([]byte, l)
			n, err := r.Read(bytes)
			if err != nil {
				return HttpRequest{}, err
			}
			if n == l {
				content = string(bytes)
			}
		}
	}
	return HttpRequest{
		Method: method,
		Path: path,
		Proto: protocol,
		Header: header,
		Content: content,
	}, nil
}

func (req HttpRequest) Protocol() Proto {
	return req.Proto
}

func (req HttpRequest) Sstp() (SstpRequest, error) {
	r := bufio.NewReader(strings.NewReader(req.Content))
	bytes, _, err := r.ReadLine()
	if err != nil {
		return SstpRequest{}, err
	}
	line := string(bytes)
	line = strings.TrimRight(line, "\r\n")
	s := strings.Split(line, " ")
	method, _, protocol, ok := parseProtocol(s)
	if !ok || protocol.Name != "SSTP" {
		return SstpRequest{}, errors.New("invalid status line")
	}
	return parseSSTP(r, method, protocol)
}

func (req HttpRequest) KeepAlive() bool {
	keep_alive, ok := req.Header["Connection"]
	if ok && keep_alive == "keep-alive" {
		return true
	}
	return false
}

func parseSSTP(r *bufio.Reader, method string, protocol Proto) (SstpRequest, error) {
	header := make(map[string]string)
	script := make(map[string]string)
	reference := make([]string, 0)
	ifghost := ""
	for {
		bytes, _, err := r.ReadLine()
		if err != nil {
			return SstpRequest{}, err
		}
		line := string(bytes)
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		s := strings.SplitN(line, ":", 2)
		if (len(s) < 2) {
			continue
		}
		k := strings.TrimSpace(s[0])
		v := strings.TrimSpace(s[1])
		if k == "Script" {
			script[ifghost] = v
		} else if k == "IfGhost" {
			ifghost = v
		} else if strings.HasPrefix(k, "Reference") {
			n, err := strconv.Atoi(k[9:])
			if err != nil {
				header[k] = v
			} else {
				for i := len(reference); i <= n; i++ {
					reference = append(reference, "")
				}
				reference[n] = v
			}
		} else {
			header[k] = v
		}
	}
	return SstpRequest{
		Method: method,
		Proto: protocol,
		Header: header,
		Script: script,
		Reference: reference,
	}, nil
}

func (req SstpRequest) Protocol() Proto {
	return req.Proto
}

func (req SstpRequest) Sstp() (SstpRequest, error) {
	return req, nil
}

func (req SstpRequest) KeepAlive() bool {
	keep_alive, ok := req.Header["Connection"]
	if ok && keep_alive == "keep-alive" {
		return true
	}
	return false
}
