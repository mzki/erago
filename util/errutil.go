// Package errutil provides utilty of errors.
package util

import (
	"fmt"
	"io"
)

// errWriter is wraper of Writer with internal Error.
// if call Write(), error is remaindered in internal,
// and trailing Write() is not executed.
type errWriter struct {
	w   io.Writer
	err error
}

// construct with io.Writer.
func NewErrWriter(w io.Writer) *errWriter { return &errWriter{w: w} }

// return internal error.
func (ew errWriter) Err() error { return ew.err }

// Write binds error of Write() to internal.
// if internal err is not nil, after write process is ignored
func (ew *errWriter) Write(p []byte) (int, error) {
	if ew.Err() != nil {
		return 0, nil // do nothing
	}
	var b int
	b, ew.err = ew.w.Write(p)
	return b, nil
}

// errReader is wrapper of io.Reader with internal error.
type errReader struct {
	r   io.Reader
	err error
}

func NewErrReader(r io.Reader) *errReader { return &errReader{r: r} }

func (er errReader) Err() error { return er.err }

func (er *errReader) Read(p []byte) (int, error) {
	if er.Err() != nil {
		return 0, nil // do nothing
	}
	var b int
	b, er.err = er.r.Read(p)
	return b, nil
}

// multiError has multipule errors in internal.
type multiError struct {
	errs []error
}

// Constract with no argument.
func NewMultiErr() *multiError {
	return &multiError{errs: make([]error, 0, 4)}
}

// Add given error into Internal.
// if error is nil, no action for internal errors.
func (me *multiError) Add(err error) {
	if err == nil {
		return
	}
	me.errs = append(me.errs, err)
}

// Err returns internal errors joined to one error.
// if internal errors is nothing, return nil.
func (me *multiError) Err() error {
	if len(me.errs) == 0 {
		return nil
	}
	str := "multiple errors:\n"
	for i, err := range me.errs {
		str += fmt.Sprintf("  %v. err: %v\n", i, err)
	}
	return fmt.Errorf("%v", str)
}
