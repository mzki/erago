// Package errutil provides utilty of errors.
package errutil

import (
	"fmt"
	"io"
)

// Writer is wraper of io.Writer with internal Error.
// if call Write(), error is remaindered in internal,
// and trailing Write() is not executed.
type Writer struct {
	w   io.Writer
	err error
}

// construct with io.Writer.
func NewErrWriter(w io.Writer) *Writer { return &Writer{w: w} }

// return internal error.
func (ew Writer) Err() error { return ew.err }

// Write binds error of Write() to internal.
// if internal err is not nil, after write process is ignored
func (ew *Writer) Write(p []byte) (int, error) {
	if ew.Err() != nil {
		return 0, nil // do nothing
	}
	var b int
	b, ew.err = ew.w.Write(p)
	return b, nil
}

// Reader is wrapper of io.Reader with internal error.
type Reader struct {
	r   io.Reader
	err error
}

func NewErrReader(r io.Reader) *Reader { return &Reader{r: r} }

func (er Reader) Err() error { return er.err }

func (er *Reader) Read(p []byte) (int, error) {
	if er.Err() != nil {
		return 0, nil // do nothing
	}
	var b int
	b, er.err = er.r.Read(p)
	return b, nil
}

// MultiError has multipule errors in internal and
// can show all of these
type MultiError struct {
	errs []error
}

// Constract with no argument.
func NewMultiError() *MultiError {
	return &MultiError{errs: make([]error, 0, 4)}
}

// Add given error into Internal.
// if error is nil, no action for internal errors.
func (me *MultiError) Add(err error) {
	if err == nil {
		return
	}
	me.errs = append(me.errs, err)
}

// Err returns internal errors joined to one error.
// if internal errors is nothing, return nil.
func (me *MultiError) Err() error {
	if len(me.errs) == 0 {
		return nil
	}
	str := "multiple errors:\n"
	for i, err := range me.errs {
		str += fmt.Sprintf("  %v. err: %v\n", i, err)
	}
	return fmt.Errorf("%v", str)
}
