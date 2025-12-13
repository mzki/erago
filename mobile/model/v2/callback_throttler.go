package model

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mzki/erago/uiadapter"
	"github.com/mzki/erago/util/log"
	"github.com/mzki/erago/view/exp/text/pubdata"
	"github.com/mzki/erago/view/exp/text/publisher"
)

type paragraphListBinaryEncoderFunc func(*pubdata.ParagraphList) ([]byte, error)

type CallbackThrottler struct {
	ctx           context.Context
	duration      time.Duration
	ticker        *time.Ticker
	syncCh        chan *publisher.Message
	closeCh       chan struct{}
	doneCh        chan struct{}
	closeOnce     *sync.Once
	messageLooper *publisher.MessageLooper
	currentSyncID atomic.Uint64
	asyncErr      *asyncError

	pendingEvents *pendingEvents
	ui            UI
	encoder       paragraphListBinaryEncoderFunc
}

func NewCallbackThrottler(ctx context.Context, d time.Duration, ui UI, encoder paragraphListBinaryEncoderFunc) *CallbackThrottler {
	if d <= 0 {
		panic("duration must be grather than zero")
	}
	return &CallbackThrottler{
		ctx:           ctx,
		duration:      d,
		ticker:        time.NewTicker(d),
		syncCh:        make(chan *publisher.Message),
		closeCh:       make(chan struct{}),
		doneCh:        make(chan struct{}),
		closeOnce:     &sync.Once{},
		messageLooper: publisher.NewMessageLooper(ctx),
		currentSyncID: atomic.Uint64{},
		asyncErr:      newAsyncError(),
		pendingEvents: newPendingEvents(),
		ui:            ui,
		encoder:       encoder,
	}
}

func (thr *CallbackThrottler) throttleLoop(ctx context.Context) {
	handleAsyncErr := func(newErr error) {
		thr.asyncErr.Set(errors.Join(
			thr.asyncErr.Err(),
			newErr,
		))
		thr.messageLooper.Close()
	}
	defer close(thr.doneCh)
	for {
		select {
		case <-ctx.Done():
			return
		case <-thr.closeCh:
			return
		case msg := <-thr.syncCh:
			err := thr.messageLooper.Send(thr.ctx, msg)
			if err != nil {
				handleAsyncErr(fmt.Errorf("throttleLoop failed to send sync message: %w", err))
				return // to quit loop
			}
			thr.ticker.Reset(thr.duration) // to defer to publish event since it was already triggred by extrenal.
		case <-thr.ticker.C:
			msg := thr.createAsyncTask(func() {
				err := thr.pendingEvents.publish(thr.ui, thr.encoder)
				if err != nil {
					handleAsyncErr(fmt.Errorf("throttleLoop failed to publish on tick event: %w", err))
				}
			})
			err := thr.messageLooper.Send(thr.ctx, msg)
			if err != nil {
				handleAsyncErr(fmt.Errorf("throttleLoop failed to send publish task: %w", err))
				return // quit throttle loop
			}
		}
	}
}

func (thr *CallbackThrottler) StartThrottle() {
	go thr.throttleLoop(thr.ctx)
}

func (thr *CallbackThrottler) Close() error {
	thr.closeOnce.Do(func() {
		close(thr.closeCh)
		// TODO: use sync task to wait for close
		thr.messageLooper.Close()
		thr.ticker.Stop()
	})
	select {
	case <-thr.doneCh:
		return nil
	case <-time.After(3 * time.Second):
		return fmt.Errorf("CallbackThrottler.Close: timeout to wait end throttle thread")
	}
}

func (thr *CallbackThrottler) IsClosed() bool {
	select {
	// TODO: need to distinguish close and done?
	case <-thr.closeCh:
		return true // close requested
	case <-thr.doneCh:
		return true // close confirmed
	default:
		return false
	}
}

func (thr *CallbackThrottler) createSyncTask(task func()) *publisher.Message {
	newID := thr.currentSyncID.Add(1)
	taskid := publisher.MessageID(newID)
	return &publisher.Message{
		ID:   taskid,
		Type: publisher.MessageSyncTask,
		Task: task,
	}
}

func (thr *CallbackThrottler) createAsyncTask(task func()) *publisher.Message {
	return &publisher.Message{
		ID:   publisher.DefaultMessageID,
		Type: publisher.MessageAsyncTask,
		Task: task,
	}
}

func (thr *CallbackThrottler) handleLooperError(err error) (ret error) {
	if errors.Is(err, publisher.ErrMessageLooperClosed) {
		ret = errors.Join(err, thr.asyncErr.Err())
	} else if errors.Is(err, context.Canceled) {
		// cancelled means application ends. notfity ErrorPipelineClosed so that application can know this errpr can be ignored
		ret = errors.Join(err, uiadapter.ErrorPipelineClosed)
	} else {
		ret = err
	}
	return
}

type asyncError struct {
	mu    *sync.Mutex
	value error
}

func newAsyncError() *asyncError {
	return &asyncError{
		mu:    new(sync.Mutex),
		value: nil,
	}
}

func (e *asyncError) Set(err error) {
	e.mu.Lock()
	e.value = err
	e.mu.Unlock()
}

func (e *asyncError) Err() error {
	var err error
	e.mu.Lock()
	err = e.value
	e.mu.Unlock()
	return err
}

// ---------- Publiser.Callback APIs. -----------------------------------------

// OnPublish is called when Paragraph is fixed by hard return (\n).
func (thr *CallbackThrottler) OnPublish(p *pubdata.Paragraph) error {
	msg := thr.createAsyncTask(func() {
		thr.pendingEvents.appendAddParagraph(p)
	})
	err := thr.messageLooper.Send(thr.ctx, msg)
	return thr.handleLooperError(err)
}

// OnPublishTemporary is called when Paragraph is NOT fixed yet by hard return(\n),
// but required to show on UI.
func (thr *CallbackThrottler) OnPublishTemporary(p *pubdata.Paragraph) error {
	return thr.OnPublish(p)
}

// OnRemove is called when game thread requests to remove (N-1)-paragraphs which have been fixed
// by calling OnPublish and also temporal Paragraph by calling OnPublishTemporary, thus to remove N-paragraphs.
func (thr *CallbackThrottler) OnRemove(nParagraph int) error {
	msg := thr.createAsyncTask(func() {
		thr.pendingEvents.appendRemoveParagraph(nParagraph)
	})
	err := thr.messageLooper.Send(thr.ctx, msg)
	return thr.handleLooperError(err)
}

// OnRemoveAll is called when game thread requests to remove all paragraphs which have been fixed
// by calling OnPublish and also temporal Paragraph by calling OnPublishTemporary.
func (thr *CallbackThrottler) OnRemoveAll() error {
	msg := thr.createAsyncTask(func() {
		thr.pendingEvents.appendRemoveParagraphAll()
	})
	err := thr.messageLooper.Send(thr.ctx, msg)
	return thr.handleLooperError(err)
}

// OnSync is called when game thread requests to synchronize pending update events to UI, such as publised paragaph and remove request.
func (thr *CallbackThrottler) OnSync() error {
	var taskErr error
	msg := thr.createSyncTask(func() {
		taskErr = thr.pendingEvents.publish(thr.ui, thr.encoder)
		if taskErr != nil {
			// taskError will be retuned to caller at end of OnSync()
			closeErr := thr.Close()
			taskErr = errors.Join(taskErr, closeErr)
		}
	})
	ctx, cancel := context.WithTimeout(thr.ctx, 5*time.Second)
	defer cancel()
	msg.SyncCtx = ctx

	select {
	case thr.syncCh <- msg:
		// msg will be sent to messageLooper in throttleLoop
	case <-ctx.Done():
		return ctx.Err()
	case <-thr.doneCh:
		return uiadapter.ErrorPipelineClosed
	}

	err := thr.messageLooper.WaitDone(ctx, msg.ID)
	if err != nil {
		return thr.handleLooperError(err)
	}
	return taskErr
}

// ---------- UiAdapter.RequestChangedListener APIs. -----------------------------------------

// implement uiadapter.RequestChangedListener
func (thr *CallbackThrottler) OnRequestChanged(req uiadapter.InputRequestType) {
	msg := thr.createAsyncTask(func() {
		thr.pendingEvents.appendInputRequest(req)
	})
	err := thr.messageLooper.Send(thr.ctx, msg)
	if err != nil {
		// TODO: the interface do not have returning error just logging.
		handledErr := thr.handleLooperError(err)
		log.Debugf("OnRequestChanged failed to send async task. Maybe application ends or exception happened at anywhere? err: %v", handledErr)
	}
}

// ---------- PendingEvents  -----------------------------------------

type pendingEvents struct {
	events []peEvent
}

func newPendingEvents() *pendingEvents {
	return &pendingEvents{
		events: make([]peEvent, 0, 16),
	}
}

func (pe *pendingEvents) appendAddParagraph(p *pubdata.Paragraph) {
	if len(pe.events) == 0 {
		paragraphs := make([]*pubdata.Paragraph, 0, 4)
		paragraphs = append(paragraphs, p)
		pe.events = append(pe.events, (*peEventParagraphList)(&paragraphs))
		return
	}

	last := pe.events[len(pe.events)-1]
	merged := last.MergeParagraph(p)
	if !merged {
		// append event newly.
		paragraphs := make([]*pubdata.Paragraph, 0, 4)
		paragraphs = append(paragraphs, p)
		pe.events = append(pe.events, (*peEventParagraphList)(&paragraphs))
	}
}

func (pe *pendingEvents) appendRemoveParagraph(nRemove int) {
	if len(pe.events) == 0 {
		pe.events = append(pe.events, (*peEventRemoveCount)(&nRemove))
		return
	}

	var (
		merged      bool
		nRestRemove = nRemove
	)
	// atLast is loop for 1..N. First element should be remained since UI side state can not be guranteed from pendingEvent side.
	for atLast := len(pe.events) - 1; atLast > 0; atLast = atLast - 1 {
		last := pe.events[atLast]
		merged, nRestRemove = last.MergeRemoveCount(nRestRemove)

		switch {
		// unmerged case have 1 pattern:
		// 1. both last element and remove request still exist
		case !merged:
			pe.events = append(pe.events[:atLast+1], (*peEventRemoveCount)(&nRestRemove))
			return
		// merged case have 3 patterns:
		// 1. last element > remove request. --> remove request is gone.
		case merged && !last.IsEmpty():
			pe.events = pe.events[:atLast+1]
			return
		// 2. last element == remove request. --> both last element and remove request are gone.
		case merged && last.IsEmpty() && nRestRemove == 0:
			pe.events = pe.events[:atLast]
			return
		// 3. last element < remove request --> last element is gone
		case merged && last.IsEmpty() && nRestRemove > 0:
			continue
		default:
			panic(fmt.Errorf("unknown pattern for last element and remove request: merged %v, last element %#+v, nRestRemove %v", merged, last, nRestRemove))
		}
	}
	// loop completed --> atLast will be 0, meaning every events except at first are gone, but remove request remains
	pe.events = append(pe.events[:1], (*peEventRemoveCount)(&nRestRemove))
}

func (pe *pendingEvents) appendRemoveParagraphAll() {
	atLast := len(pe.events) - 1
	for ; atLast >= 0; atLast = atLast - 1 {
		last := pe.events[atLast]
		if last.Type() == peEventTypeInputRequest {
			break
		}
	}
	pe.events = pe.events[:atLast+1] // clear all of paragraph addition and remove count except input request since remove ALL.
	// taking into account there are paragraphs already published at UI side. need to remove those by notifying remove all event.
	pe.events = append(pe.events, peEventRemoveAll{})
}

func (pe *pendingEvents) appendInputRequest(r uiadapter.InputRequestType) {
	lastAt := len(pe.events) - 1
	if lastAt < 0 {
		pe.events = append(pe.events, (*peEventInputRequest)(&r))
		return
	}

	merged := pe.events[lastAt].MergeInputRequest(r)
	if merged {
		pe.events = pe.events[:lastAt+1]
		return
	}

	// Input event change is important for ui side state change. Need to remain every change during pending events.
	// there is no point to merge except last. just append it at last
	pe.events = append(pe.events, (*peEventInputRequest)(&r))
}

func (pe *pendingEvents) publish(ui UI, encoder paragraphListBinaryEncoderFunc) error {
	eventError := func(eventName, actName string, err error) error {
		return fmt.Errorf("publish %s: %s: error: %w", eventName, actName, err)
	}
	for _, ev := range pe.events {
		var eventName string = "unknown"
		switch ev.Type() {
		case peEventTypeParagraphList:
			eventName = "ParagraphList"
			plist := &pubdata.ParagraphList{
				Paragraphs: ev.(*peEventParagraphList).getRawData(),
			}
			bs, err := encoder(plist)
			if err != nil {
				return eventError(eventName, "encode", err)
			}
			err = ui.OnPublishBytes(bs)
			if err != nil {
				return eventError(eventName, "OnPublishBytes", err)
			}
		case peEventTypeRemoveCount:
			eventName = "RemoveCount"
			err := ui.OnRemove(ev.(*peEventRemoveCount).getRawData())
			if err != nil {
				return eventError(eventName, "OnRemove", err)
			}
		case peEventTypeRemoveAll:
			eventName = "RemoveAll"
			err := ui.OnRemoveAll()
			if err != nil {
				return eventError(eventName, "OnRemoveAll", err)
			}
		case peEventTypeInputRequest:
			eventName = "InputRequest"
			switch r := ev.(*peEventInputRequest).getRawData(); r {
			case uiadapter.InputRequestCommand, uiadapter.InputRequestRawInput:
				ui.OnCommandRequested()
			case uiadapter.InputRequestInput:
				ui.OnInputRequested()
			case uiadapter.InputRequestNone:
				ui.OnInputRequestClosed()
			}
		}
	}
	// clear publushed events.
	pe.events = pe.events[:0]
	return nil
}

type (
	peEventParagraphList []*pubdata.Paragraph
	peEventRemoveCount   int
	peEventRemoveAll     struct{}
	peEventInputRequest  uiadapter.InputRequestType

	peEventType int

	peEvent interface {
		Type() peEventType
		IsEmpty() bool
		MergeParagraph(p *pubdata.Paragraph) (merged bool)
		MergeRemoveCount(nCount int) (merged bool, restNCount int)
		MergeInputRequest(r uiadapter.InputRequestType) (merged bool)
	}
)

const (
	peEventTypeNone peEventType = iota
	peEventTypeParagraphList
	peEventTypeRemoveCount
	peEventTypeRemoveAll
	peEventTypeInputRequest
)

func (peEventParagraphList) Type() peEventType    { return peEventTypeParagraphList }
func (plist *peEventParagraphList) IsEmpty() bool { return len(plist.getRawData()) == 0 }

func (plist *peEventParagraphList) getRawData() []*pubdata.Paragraph {
	return ([]*pubdata.Paragraph)(*plist)
}
func (plist *peEventParagraphList) setRawData(v []*pubdata.Paragraph) {
	*plist = peEventParagraphList(v)
}

func (plist *peEventParagraphList) MergeParagraph(p *pubdata.Paragraph) (merged bool) {
	ps := plist.getRawData()
	if len(ps) > 0 && ps[len(ps)-1].Id == p.Id {
		ps[len(ps)-1] = p // just replacement
	} else {
		ps = append(ps, p)
	}
	plist.setRawData(ps)
	return true
}

func (plist *peEventParagraphList) MergeRemoveCount(nCount int) (merged bool, restNCount int) {
	ps := plist.getRawData()
	if len(ps) == 0 {
		return true, nCount
	}

	// nCount actually exclude the last unfixed Paragragh which is .fixed = false and will be updated later.
	// nCount + 1 include the last unfixed Paragraph.
	nCountPlus1 := nCount + 1
	if lastFixed := ps[len(ps)-1].Fixed; lastFixed {
		// unfixed paragraph is not yet published. We can just remove fixed paragraphs by nCount
		if len(ps) >= nCount {
			restNCount = 0
			ps = ps[:len(ps)-nCount]
		} else {
			restNCount = nCount - len(ps)
			ps = ps[:0]
		}
	} else {
		// unfixed paragraph is published. We can remove fixed paragraph + an unfixed paragraph by nCountPlus1
		if len(ps) >= nCountPlus1 {
			last := ps[len(ps)-1]
			makeEmptyParagraph(last)
			ps = ps[:len(ps)-nCountPlus1]
			ps = append(ps, last) // remains empty unfixed paragragh at last.
			restNCount = 0
		} else {
			// unfixed paragraph is completely removed. it will not be published to ui side.
			// This case is complex and not able to handle correctly in pending event side. Remaing remove count as is.
			//
			// UI:
			//   Paragraph{ID=1, fixed=true, ...}
			// PendingEvent:
			//   Paragraph{ID=2, fixed=true, ...}
			//   Paragraph{ID=3, fixed=true, ...}
			//   Paragraph{ID=4, fixed=false, ...}
			// Remove: 4 (nCount always minus 1 for un-fixed paragraph. So nCount = 3 here)
			restNCount = nCount
			return false, restNCount
		}
	}
	plist.setRawData(ps)
	return true, restNCount
}

func makeEmptyParagraph(p *pubdata.Paragraph) {
	p.Lines = p.Lines[:1]
	l := p.Lines[0]
	l.RuneWidth = 0
	l.Boxes = l.Boxes[:1]
	b := l.Boxes[0]
	b.ContentType = pubdata.ContentType_CONTENT_TYPE_SPACE
	b.Data = &pubdata.Box_SpaceData{SpaceData: &pubdata.SpaceData{}}
	b.RuneWidth = 0
	b.LineCountHint = 1
}

func (peEventParagraphList) MergeInputRequest(r uiadapter.InputRequestType) bool {
	return false /* never */
}

func (peEventRemoveCount) Type() peEventType { return peEventTypeRemoveCount }
func (peEventRemoveCount) IsEmpty() bool {
	return false /* never. since there is the case of unfixed last paragraph */
}

func (ev *peEventRemoveCount) getRawData() int  { return int(*ev) }
func (ev *peEventRemoveCount) setRawData(v int) { *ev = (peEventRemoveCount)(v) }

func (peEventRemoveCount) MergeParagraph(p *pubdata.Paragraph) (merged bool) { return false /* never */ }

func (rmCount *peEventRemoveCount) MergeRemoveCount(nCount int) (merged bool, restNCount int) {
	selfCount := rmCount.getRawData()
	rmCount.setRawData(selfCount + nCount)
	return true, 0
}
func (peEventRemoveCount) MergeInputRequest(r uiadapter.InputRequestType) bool {
	return false /* never */
}

func (peEventRemoveAll) Type() peEventType { return peEventTypeRemoveAll }
func (peEventRemoveAll) IsEmpty() bool     { return false /* never */ }

func (peEventRemoveAll) MergeParagraph(p *pubdata.Paragraph) (merged bool) { return false /* never */ }
func (peEventRemoveAll) MergeRemoveCount(nCount int) (merged bool, restNCount int) {
	return true, 0 /* consumed nCount */
}
func (peEventRemoveAll) MergeInputRequest(r uiadapter.InputRequestType) bool { return false /* never */ }

func (peEventInputRequest) Type() peEventType { return peEventTypeInputRequest }
func (ev peEventInputRequest) IsEmpty() bool  { return false /* never */ }
func (ev peEventInputRequest) getRawData() uiadapter.InputRequestType {
	return uiadapter.InputRequestType(ev)
}

// func (ev *peEventInputRequest) setRawData(other uiadapter.InputRequestType) {
// 	*ev = peEventInputRequest(other)
// }

func (peEventInputRequest) MergeParagraph(p *pubdata.Paragraph) (merged bool) {
	return false /* never */
}
func (peEventInputRequest) MergeRemoveCount(nCount int) (merged bool, restNCount int) {
	return false, nCount /* never */
}

func (ev *peEventInputRequest) MergeInputRequest(r uiadapter.InputRequestType) (merged bool) {
	self := ev.getRawData()
	return self == r // just merge for same event only.
	// since input event is important for ui side handling. need precise event stream even if
	// event frequency is high load.
}
