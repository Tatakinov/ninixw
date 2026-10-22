package utils

import (
	"bufio"
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
