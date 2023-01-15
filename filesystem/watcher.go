package filesystem

import (
	"fmt"
	"time"

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

	events chan WatchEvent
	errors chan error
	done   chan struct{}
}

func (wi *watcherImpl) eventLoop() {
	// The backend fsnotify fires same event twice depending on specific applications.
	// This event loop puts multiple events toghether into one in a pendingDuration and
	// publish it later to guranantee event fires exactly once.
	// See https://github.com/fsnotify/fsnotify/issues/122
	defer func() {
		close(wi.events)
		close(wi.errors)
	}()
	const pendingDuration = 100 * time.Millisecond
	pendingTimer := time.NewTimer(pendingDuration)
	<-pendingTimer.C // consume first event, which is not related to anywhere.
	pendingEvents := make(map[WatchEvent]bool)
	for {
		select {
		case <-wi.done:
			return
		case ev, ok := <-wi.w.Events:
			if !ok {
				return
			}
			pendingEvents[ev] = true
			// reset timer to fire event later.
			if !pendingTimer.Stop() {
				// needs select here since sometimes there is a case that
				// pendingTimer.C drained and Stop() returns false, causes infinite blocking here.
				select {
				case <-pendingTimer.C:
				default:
				}
			}
			pendingTimer.Reset(pendingDuration)
		case err, ok := <-wi.w.Errors:
			if !ok {
				return
			}
			select {
			case wi.errors <- err:
			case <-wi.done:
			}
		case <-pendingTimer.C:
			for ev := range pendingEvents {
				select {
				case wi.events <- ev:
				case <-wi.done:
				}
				delete(pendingEvents, ev) // clear consumed event
			}
		}
	}
}

func (wi *watcherImpl) Close() error {
	select {
	case <-wi.done:
	default:
		close(wi.done)
	}
	return wi.w.Close()
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

func (wi *watcherImpl) Events() <-chan WatchEvent { return wi.events }
func (wi *watcherImpl) Errors() <-chan error      { return wi.errors }

func newWatcher(pr PathResolver) (Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("NewWatcher failed by backend fsnotify.NewWatcher(): %w", err)
	}
	wi := &watcherImpl{
		w:            w,
		pathResolver: pr,
		events:       make(chan WatchEvent),
		errors:       make(chan error),
		done:         make(chan struct{}),
	}
	go wi.eventLoop()
	return wi, nil
}
