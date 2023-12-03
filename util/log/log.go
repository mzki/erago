// package log defines strict logger types, which is referenced from
// https://dave.cheney.net/2015/11/05/lets-talk-about-logging.
package log

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

const (
	// These are specifies logging level.
	// InfoLevel outputs only calling Info*.
	// DebugLevel outputs all of outputting call, Info* and Debug*.
	InfoLevel  = iota // output only Info*
	DebugLevel        // output all, Info* and Debug*
)

// DebugPrefix is 2nd prefix of outputting text when call Debug*.
// The most left prefix is which is used for SetPrefix or New(.., prefix, ...).
// So DebugPrefix places right by the most left prefix and left by the outputting text.
const DebugPrefix = "DEBUG: "

// ErrWriteDiscadedByLevel indicates log output is discarded by different level, e.g. Debug() with info level.
var ErrOutputDiscardedByLevel = errors.New("log output discarded by different log level")

// Simple logger which has only 2 levels, info and debug only.
// Its output error is not retruned for convinient purpose. the latest output error
// is recorded internally and can be retrived later from Err() API.
type Logger struct {
	logger *log.Logger

	mu          sync.Mutex
	level       int // output level, under the mutex.
	internalErr error
}

func (l *Logger) internalInfo(calldepth int, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	err := l.logger.Output(calldepth, msg)
	l.internalErr = err
}

func (l *Logger) Info(v ...interface{})                 { l.internalInfo(3, fmt.Sprint(v...)) }
func (l *Logger) Infoln(v ...interface{})               { l.internalInfo(3, fmt.Sprintln(v...)) }
func (l *Logger) Infof(format string, v ...interface{}) { l.internalInfo(3, fmt.Sprintf(format, v...)) }

func (l *Logger) internalDebug(calldepth int, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.level < DebugLevel {
		l.internalErr = ErrOutputDiscardedByLevel
		return
	}
	err := l.logger.Output(calldepth, DebugPrefix+msg)
	l.internalErr = err
}

func (l *Logger) Debug(v ...interface{})   { l.internalDebug(3, fmt.Sprint(v...)) }
func (l *Logger) Debugln(v ...interface{}) { l.internalDebug(3, fmt.Sprintln(v...)) }
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.internalDebug(3, fmt.Sprintf(format, v...))
}

func (l *Logger) SetOutput(w io.Writer) {
	l.logger.SetOutput(w)
}

// set logging level.
func (l *Logger) SetLevel(level int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// return current logging level.
func (l *Logger) Level() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// same as standard package's log
func (l *Logger) SetFlags(flag int) {
	l.logger.SetFlags(flag)
}

// same as standard package's log
func (l *Logger) Flags() int {
	return l.logger.Flags()
}

// same as standard package's log
func (l *Logger) SetPrefix(prefix string) {
	l.logger.SetPrefix(prefix)
}

// same as standard package's log
func (l *Logger) Prefix() string {
	return l.logger.Prefix()
}

// Err returns last internal erorr in logger.
// Even If the internal error is occured pastly, but last time succeeded and no error (nil) then internal error is
// replaced by last time result (nil).
//
//	logger.Info("1") --> something error
//	logger.Info("2") --> no erorr
//	logger.Err() --> nil
//
// If discarding output message by log level, for example Debug() is discarded with info level,
// Err() should returns ErrOutputDiscardedByLevel.
func (l *Logger) Err() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.internalErr
}

const (
	// These flags are same as log package's.
	Ldate         = log.Ldate         // the date in the local time zone: 2009/01/23
	Ltime         = log.Ltime         // the time in the local time zone: 01:23:23
	Lmicroseconds = log.Lmicroseconds // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile     = log.Llongfile     // full file name and line number: /a/b/c/d.go:23
	Lshortfile    = log.Lshortfile    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC          = log.LUTC          // if Ldate or Ltime is set, use UTC rather than the local time zone
	LstdFlags     = log.LstdFlags     // initial values for the standard logger
)

// construct new Logger. default output level is InfoLevel.
func New(out io.Writer, prefix string, flag int) *Logger {
	return &Logger{
		logger:      log.New(out, prefix, flag),
		level:       InfoLevel,
		internalErr: nil,
	}
}

var std = New(os.Stdout, "", LstdFlags)

func Info(v ...interface{}) {
	std.internalInfo(3, fmt.Sprint(v...))
}

func Infoln(v ...interface{}) {
	std.internalInfo(3, fmt.Sprintln(v...))
}

func Infof(format string, v ...interface{}) {
	std.internalInfo(3, fmt.Sprintf(format, v...))
}

func Debug(v ...interface{}) {
	std.internalDebug(3, fmt.Sprint(v...))
}

func Debugln(v ...interface{}) {
	std.internalDebug(3, fmt.Sprintln(v...))
}

func Debugf(format string, v ...interface{}) {
	std.internalDebug(3, fmt.Sprintf(format, v...))
}

func SetOutput(w io.Writer) {
	std.SetOutput(w)
}

// set logging level to default logger.
func SetLevel(level int) {
	std.SetLevel(level)
}

// return current logging level for default logger.
func Level() int {
	return std.Level()
}

func SetFlags(flag int) {
	std.SetFlags(flag)
}

func Flags() int {
	return std.Flags()
}

func SetPrefix(prefix string) {
	std.SetPrefix(prefix)
}

func Prefix() string {
	return std.Prefix()
}

func Err() error {
	return std.Err()
}

// SilentWriter implements io.Writer.
// it writes no content and return no error so that
// any writing is ignored.
type SilentWriter struct{}

func (SilentWriter) Write([]byte) (int, error) {
	return 0, nil
}

// https://go-review.googlesource.com/c/go/+/319593/12/src/internal/iointernal/limited_writer.go

// LimitWriter returns a Writer that writes to w
// but stops with EOF after n bytes.
// The underlying implementation is a *LimitedWriter.
func LimitWriter(w io.Writer, n int64) io.Writer { return &LimitedWriter{w, n} }

// A LimitedWriter writes to W but limits the amount of
// data returned to just N bytes. Each call to Write
// updates N to reflect the new amount remaining.
// Read returns EOF when N <= 0 or when the underlying W returns EOF.
type LimitedWriter struct {
	W io.Writer // underlying writer
	N int64     // max bytes remaining
}

func (l *LimitedWriter) Write(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.W.Write(p)
	l.N -= int64(n)
	return
}
