package utils

import (
	"bufio"
	"io"
	"os"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func ReadLine(r *bufio.Reader) ([]byte, error) {
	buffer := make([]byte, 0)
	for {
		bytes, is_prefix, err := r.ReadLine()
		if err != nil {
			return []byte{}, err
		}
		buffer = append(buffer, bytes...)
		if !is_prefix {
			break
		}
	}
	return buffer, nil
}

func ReadSetting(p string) (map[string]string, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	reader := transform.NewReader(f, japanese.ShiftJIS.NewDecoder())
	scanner := bufio.NewScanner(reader)
	charset := "Shift_JIS"
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "charset") {
			continue
		}
		a := strings.SplitN(line, ",", 2)
		if len(a) < 2 {
			continue
		}
		charset = strings.TrimSpace(a[1])
		break
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	if (charset == "Shift_JIS") {
		reader := transform.NewReader(f, japanese.ShiftJIS.NewDecoder())
		scanner = bufio.NewScanner(reader)
	} else {
		scanner = bufio.NewScanner(f)
	}
	m := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		a := strings.SplitN(line, ",", 2)
		if len(a) < 2 {
			continue
		}
		m[strings.TrimSpace(a[0])] = strings.TrimSpace(a[1])
	}
	return m, nil
}
