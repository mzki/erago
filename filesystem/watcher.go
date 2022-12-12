package filesystem

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
)

// Export some types and values so that user need not import underlying package explicitly
type WatchEvent = fsnotify.Event
type WatchOp = fsnotify.Op

const (
	WatchOpCreate = fsnotify.Create
	WatchOpWrite  = fsnotify.Write
	WatchOpRemove = fsnotify.Remove
	WatchOpRename = fsnotify.Rename
	WatchOpChmod  = fsnotify.Chmod
)

var (
	ErrNonExistentWatch = fsnotify.ErrNonExistentWatch
	ErrEventOverflow    = fsnotify.ErrEventOverflow
)

type Watcher interface {
	Watch(filepath string) error
	UnWatch(filepath string) error
	Events() <-chan WatchEvent
	Errors() <-chan error
	Close() error
}

type watcherImpl struct {
	w            *fsnotify.Watcher
	pathResolver PathResolver
}

func (wi *watcherImpl) Watch(filepath string) error {
	p, err := wi.pathResolver.ResolvePath(filepath)
	if err != nil {
		return fmt.Errorf("failed to Watch(%s): %w", filepath, err)
	}
	return wi.w.Add(p)
}

func (wi *watcherImpl) UnWatch(filepath string) error {
	p, err := wi.pathResolver.ResolvePath(filepath)
	if err != nil {
		return fmt.Errorf("failed to UWatch(%s): %w", filepath, err)
	}
	return wi.w.Remove(p)
}

func (wi *watcherImpl) Events() <-chan WatchEvent { return wi.w.Events }
func (wi *watcherImpl) Errors() <-chan error      { return wi.w.Errors }
func (wi *watcherImpl) Close() error              { return wi.w.Close() }

func newWatcher(pr PathResolver) (Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("NewWatcher failed by backend fsnotify.NewWatcher(): %w", err)
	}
	return &watcherImpl{w, pr}, nil
}
