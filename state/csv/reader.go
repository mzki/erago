package csv

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/mzki/erago/filesystem"
)

const (
	// Configures of reading CSV files.
	// TODO: make it configurable?
	Separator = ","
	Comment   = ";"
)

// ReadFileFunc reads csv file formatted by the era manner.
// Each records in the csv file processed by the user function.
// It returns error with at which line occurs.
func ReadFileFunc(file string, f func([]string) error) error {
	fp, err := filesystem.Load(file)
	if err != nil {
		return err
	}
	defer fp.Close()

	if err := ReadFunc(fp, f); err != nil {
		return fmt.Errorf("%s: %v", file, err)
	}
	return nil
}

// ReadFunc reads csv data formatted by the era manner.
// Each records in the csv file processed by the user function.
// It returns error with at which line occurs.
func ReadFunc(r io.Reader, f func([]string) error) error {
	reader := NewReader(r)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		// run user function with record.
		if err := f(record); err != nil {
			return reader.readError(err)
		}
	}
	return nil
}

// Reader read csv file with the manner for the Comma and Comment.
type Reader struct {
	scanner *bufio.Scanner
	nline   int
	line    string
}

// NewReader creates Reader instance for io.Reader of a csv file.
func NewReader(r io.Reader) *Reader {
	return &Reader{
		scanner: bufio.NewScanner(r),
		nline:   0,
		line:    "",
	}
}

func (r *Reader) readError(cause error) error {
	return fmt.Errorf("line %d:'%v', Error: %v", r.nline, r.line, cause)
}

// Read reads a csv record from io.Reader.
// Calling it advances underlying buffer position.
// It returns a record contains some fields and error if read failed,
// except that if reader reached EOF, it returns (nil, io.EOF).
func (r *Reader) Read() ([]string, error) {
	// scanning loop because csv file may contain empty line or comment line
	for r.scanner.Scan() {
		r.nline++
		r.line = r.scanner.Text()
		// trimming trailing text at occuring comment symbol.
		if i := strings.Index(r.line, Comment); i != -1 { // ignore comment
			r.line = r.line[:i]
		}

		record := strings.Split(r.line, Separator)
		if len(record) < 2 {
			// ignore empty line and line having only 1 fields.
			// and retry scanning csv record
			continue
		}
		for i, field := range record {
			record[i] = strings.TrimSpace(field)
		}

		return record, nil
	}

	if err := r.scanner.Err(); err != nil {
		return nil, err
	}
	return nil, io.EOF
}
