// package log defines strict logger types, which is referenced from
// https://dave.cheney.net/2015/11/05/lets-talk-about-logging.
package log

import (
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

type Logger struct {
	logger *log.Logger

	mu    sync.Mutex
	level int // output level, under the mutex.
}

func (l *Logger) Info(v ...interface{}) {
	l.logger.Output(2, fmt.Sprint(v...))
}

func (l *Logger) Infoln(v ...interface{}) {
	l.logger.Output(2, fmt.Sprintln(v...))
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.logger.Output(2, fmt.Sprintf(format, v...))
}

func (l *Logger) Debug(v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.level < DebugLevel {
		return
	}
	l.logger.Output(2, DebugPrefix+fmt.Sprint(v...))
}

func (l *Logger) Debugln(v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.level < DebugLevel {
		return
	}
	l.logger.Output(2, DebugPrefix+fmt.Sprintln(v...))
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.level < DebugLevel {
		return
	}
	l.logger.Output(2, DebugPrefix+fmt.Sprintf(format, v...))
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
		logger: log.New(out, prefix, flag),
		level:  InfoLevel,
	}
}

var std = New(os.Stdout, "", LstdFlags)

func Info(v ...interface{}) {
	std.logger.Output(2, fmt.Sprint(v...))
}

func Infoln(v ...interface{}) {
	std.logger.Output(2, fmt.Sprintln(v...))
}

func Infof(format string, v ...interface{}) {
	std.logger.Output(2, fmt.Sprintf(format, v...))
}

func Debug(v ...interface{}) {
	std.mu.Lock()
	defer std.mu.Unlock()
	if std.level < DebugLevel {
		return
	}
	std.logger.Output(2, DebugPrefix+fmt.Sprint(v...))
}

func Debugln(v ...interface{}) {
	std.mu.Lock()
	defer std.mu.Unlock()
	if std.level < DebugLevel {
		return
	}
	std.logger.Output(2, DebugPrefix+fmt.Sprintln(v...))
}

func Debugf(format string, v ...interface{}) {
	std.mu.Lock()
	defer std.mu.Unlock()
	if std.level < DebugLevel {
		return
	}
	std.logger.Output(2, DebugPrefix+fmt.Sprintf(format, v...))
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

// SilentWriter implements io.Writer.
// it writes no content and return no error so that
// any writing is ignored.
type SilentWriter struct{}

func (SilentWriter) Write([]byte) (int, error) {
	return 0, nil
}
