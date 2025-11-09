package publisher

import (
	"context"
	"errors"
	"sync"
)

// MesssageType indicates how does the message be handled.
type MessageType uint8

const (
	MessageTypeTask  MessageType = 0
	MessageAsyncTask             = MessageTypeTask + iota // same as MessageTypeTask
	MessageSyncTask
)

// MessageID indicates a unique ID for the Message.
// This is used by SyncTask to identify which task is waited for its done.
type MessageID uint64

const (
	DefaultMessageID MessageID = 0
)

// Message is task primitive executed in a event loop.
// Only Task should be set, others are optional.
type Message struct {
	ID      MessageID
	Type    MessageType
	Task    func()
	SyncCtx context.Context // used for canceling sync task completion. can be nil when async task.
}

// ErrMEssageLooperClosed indicates API call is failed due to MessageLooper is already closed.
var ErrMessageLooperClosed = errors.New("MessageLooper is closed")

// MessageLooper executes Message tasks in its event loop.
type MessageLooper struct {
	messageCh chan *Message
	doneCh    chan MessageID
	closeCh   chan struct{}
	closeOnce sync.Once
}

// NewMessageLooper create new MessageLooper instance and start event loop internally.
// So new instance can be used immediately for Send() message.
func NewMessageLooper(ctx context.Context) *MessageLooper {
	looper := &MessageLooper{
		messageCh: make(chan *Message),
		doneCh:    make(chan MessageID),
		closeCh:   make(chan struct{}),
		closeOnce: sync.Once{},
	}
	go looper.start(ctx)
	return looper
}

func (looper *MessageLooper) start(ctx context.Context) {
message_loop:
	for {
		select {
		case <-ctx.Done():
			break message_loop
		case <-looper.closeCh:
			break message_loop
		case msg, ok := <-looper.messageCh:
			if !ok {
				break
			}
			switch msg.Type {
			case MessageTypeTask:
				fallthrough
			case MessageAsyncTask:
				if msg.Task != nil {
					msg.Task()
				}
			case MessageSyncTask:
				if msg.Task != nil {
					msg.Task()
				}
				var syncCtx context.Context = ctx
				if msg.SyncCtx != nil {
					syncCtx = msg.SyncCtx
				}
				looper.sendDone(syncCtx, msg.ID)
			}
		}
	}
}

func (looper *MessageLooper) sendDone(ctx context.Context, id MessageID) {
	select {
	case <-ctx.Done():
	case <-looper.closeCh:
	case looper.doneCh <- id:
	}
}

// Close quits internal event loop. It is goroutine safe.
func (looper *MessageLooper) Close() {
	select {
	case <-looper.closeCh:
		// already closed
		break
	default:
		looper.closeOnce.Do(func() {
			close(looper.closeCh)
			// messageCh may be accessed by write. not closed
			// close(looper.doneCh)
			// close(looper.messageCh)
		})
	}
}

func (looper *MessageLooper) Send(ctx context.Context, msg *Message) error {
	select {
	case <-looper.closeCh:
		return ErrMessageLooperClosed
	case <-ctx.Done():
		return ctx.Err()
	case looper.messageCh <- msg:
		return nil
	}
}

func (looper *MessageLooper) WaitDone(ctx context.Context, id MessageID) error {
	// wait any signal and do nothing
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-looper.closeCh:
			return ErrMessageLooperClosed
		case doneId := <-looper.doneCh:
			if doneId == id {
				return nil
			}
		}
	}
}
