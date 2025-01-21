package network

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

const (
	ARRAY_PREFIX  = '*'
	STRING_PREFIX = '$'
)

var offset = 0

func parseRESPArray(reader *bufio.Reader) ([]string, error) {
	// Initialize the offset tracker
	offset = 0

	// Read array length
	line, err := reader.ReadString('\n')
	fmt.Print("line: ", line)
	if err != nil {
		return nil, err
	}
	offset += len(line) // Count the bytes read

	if len(line) < 2 || line[0] != ARRAY_PREFIX {
		return nil, fmt.Errorf("invalid RESP array")
	}

	count, err := strconv.Atoi(strings.TrimSpace(line[1:]))
	if err != nil {
		return nil, err
	}

	args := make([]string, count)
	for i := 0; i < count; i++ {
		// Read string length
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		offset += len(line) // Count the bytes read

		if len(line) < 2 || line[0] != STRING_PREFIX {
			return nil, fmt.Errorf("invalid RESP string")
		}

		length, err := strconv.Atoi(strings.TrimSpace(line[1:]))
		if err != nil {
			return nil, err
		}

		// Read string content
		buffer := make([]byte, length+2) // +2 for \r\n
		n, err := reader.Read(buffer)
		if err != nil {
			return nil, err
		}
		offset += n // Count the bytes read
		args[i] = string(buffer[:length])
	}

	return args, nil
}

func parseRESPString(reader *bufio.Reader) (string, error) {
	// Read string length
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	if len(line) < 2 || line[0] != STRING_PREFIX {
		return "", fmt.Errorf("invalid RESP string")
	}

	length, err := strconv.Atoi(strings.TrimSpace(line[1:]))
	if err != nil {
		return "", err
	}

	// Read string content
	buffer := make([]byte, length+2) // +2 for \r\n
	_, err = reader.Read(buffer)
	if err != nil {
		return "", err
	}

	return string(buffer[:length]), nil
}
