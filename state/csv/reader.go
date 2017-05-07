package csv

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	// Configures of reading CSV files.
	// TODO: make it configurable?
	Separator = ","
	Comment   = ";"
)

// TODO: make type Reader struct{}?

// ReadFileFunc reads csv file formatted by the era manner.
// Each records in the csv file processed by the user function.
// It returns error with at which line occurs.
func ReadFileFunc(file string, f func([]string) error) error {
	fp, err := os.Open(file)
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
	nline := 0
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		nline++
		line := scanner.Text()
		// trimming trailing text at occuring comment symbol.
		if i := strings.Index(line, Comment); i != -1 { // ignore comment
			line = line[:i]
		}

		record := strings.Split(line, Separator)
		if len(record) < 2 { // ignore empty line and line having only 1 fields.
			continue
		}
		for i, field := range record {
			record[i] = strings.TrimSpace(field)
		}

		// run user function with record.
		if err := f(record); err != nil {
			return fmt.Errorf("line %d: '%v', Error: %v", nline, line, err)
		}
	}
	return scanner.Err()
}
