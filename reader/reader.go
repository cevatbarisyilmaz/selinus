package reader

import (
	"bufio"
	"os"
	"strings"
)

func ReadFile(fileName string) (*bufio.Reader, error) {
	stream, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	return bufio.NewReader(stream), nil
}

func ReadString(content string) *bufio.Reader {
	return bufio.NewReader(strings.NewReader(content))
}
