package internal

import (
	"bufio"
	"io"
	"strings"
)

type Reader struct {
	scanner *bufio.Scanner
}

// NewReader creates a new reader from a bufio.Scanner
func NewReader(scanner *bufio.Scanner) *Reader {
	return &Reader{scanner: scanner}
}

// Next reads the next token or returns an error
func (r *Reader) Next() (token string, err error) {
	scanned := r.scanner.Scan()

	if !scanned {
		return "", io.EOF
	}

	scanErr := r.scanner.Err()
	if scanErr != nil {
		return "", scanErr
	}

	return strings.TrimSpace(r.scanner.Text()), nil
}

// Current returns the current read token
func (r *Reader) Current() string {
	return strings.TrimSpace(r.scanner.Text())
}
