package publisher

import (
	"context"
	"errors"
	"image"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/mzki/erago/attribute"
	"github.com/mzki/erago/view/exp/text"
	"github.com/mzki/erago/view/exp/text/pubdata"
	"github.com/mzki/erago/width"
)

// ResetColor is imported from text.ResetColor so that
// user need not to import text package explicitly.
var ResetColor = text.ResetColor

// DefaultCachedImageSize is used as a number of cached images.
var DefaultCachedImageSize = text.DefaultCachedImageSize

type EditorOptions struct {
	// ImageFetchType indicates how does image pixels fetched by Editor.
	ImageFetchType pubdata.ImageFetchType
}

// Editor edits just one Paragraph
// It only appends new content into Frame's last.
//
// It is constructed by Frame.Editor(),
// not NewEditor() or Editor{}
//
// Multiple Editors do not share their states.
type Editor struct {
	frame     *text.Frame
	editor    *text.Editor // backend editer.
	imgLoader *ImageBytesLoader
	opt       EditorOptions

	ctx           context.Context
	looper        *MessageLooper
	currentSyncID uint64
	asyncErr      error
	asyncErrMu    *sync.Mutex

	viewParams struct {
		viewLineCount     int
		viewLineRuneWidth int
	}

	publishedCount uint64

	callback Callback
}

func NewEditor(ctx context.Context, opts ...EditorOptions) *Editor {
	opt := EditorOptions{}
	if len(opts) > 0 {
		opt = opts[0]
	}
	f := text.NewFrame(&text.FrameOptions{
		MaxParagraphs:     2,
		MaxParagraphBytes: 4 * 512, // 512 character for 4-byte.
	})
	e := &Editor{
		frame:  f,
		editor: f.Editor(),
		imgLoader: NewImageBytesLoader(
			DefaultCachedImageSize,
			// temporary width/height
			8,
			14,
			opt.ImageFetchType,
		),
		opt:           opt,
		ctx:           ctx,
		looper:        NewMessageLooper(ctx),
		currentSyncID: 0,
		asyncErr:      nil,
		asyncErrMu:    new(sync.Mutex),
		callback:      &CallbackDefault{},
	}
	return e
}

// Close closes editor APIs. After calling this, any editor's API
// return error which is errors.Is(ErrMessageLooperClosed).
// This API waits until all of pending tasks are done.
func (e *Editor) Close() error {
	msg := e.createSyncTask(func() { e.looper.Close() })
	err := e.sendAndWait(e.ctx, msg)
	if errors.Is(err, ErrMessageLooperClosed) {
		return nil // already closed.
	}
	return err
}

func (e *Editor) createSyncTask(task func()) *Message {
	taskid := MessageID(e.currentSyncID)
	e.currentSyncID++
	return &Message{
		ID:   taskid,
		Type: MessageSyncTask,
		Task: task,
	}
}

func (*Editor) createAsyncTask(task func()) *Message {
	return &Message{
		ID:   DefaultMessageID,
		Type: MessageAsyncTask,
		Task: task,
	}
}

func (e *Editor) handleLooperError(err error) (result error) {
	result = err
	if errors.Is(err, ErrMessageLooperClosed) {
		e.asyncErrMu.Lock()
		if e.asyncErr != nil {
			// Closed is caused by async error
			result = e.asyncErr
		}
		e.asyncErrMu.Unlock()
	}
	return
}

func (e *Editor) setAsyncErr(err error) {
	e.asyncErrMu.Lock()
	defer e.asyncErrMu.Unlock()
	e.asyncErr = err
}

func (e *Editor) handleAsyncErr(err error) {
	e.setAsyncErr(err)
	e.looper.Close() // looper closed by async error.
}

const sendWaitTime time.Duration = 3 * time.Second

func (e *Editor) sendAndWait(ctx context.Context, msg *Message) error {
	ctx, cancel := context.WithTimeout(ctx, sendWaitTime)
	defer cancel()
	if msg.Type != MessageSyncTask {
		panic("Message must be sync type.")
	}
	taskid := msg.ID
	if err := e.looper.Send(ctx, msg); err != nil {
		return e.handleLooperError(err)
	}
	if err := e.looper.WaitDone(ctx, taskid); err != nil {
		return e.handleLooperError(err)
	}
	return nil
}

func (e *Editor) send(ctx context.Context, msg *Message) error {
	ctx, cancel := context.WithTimeout(ctx, sendWaitTime)
	defer cancel()

	if msg.Type != MessageAsyncTask {
		panic("Message must be async type.")
	}
	if err := e.looper.Send(ctx, msg); err != nil {
		return e.handleLooperError(err)
	}
	return nil
}

// SetCallback set callback interface. This is goroutine safe,
// But some events do not call callback until reflation latency.
func (e *Editor) SetCallback(cb Callback) error {
	msg := e.createAsyncTask(func() {
		e.callback = cb
	})
	return e.send(e.ctx, msg)
}

func (e *Editor) SetViewSize(viewLineCount, viewLineRuneWidth int) error {
	msg := e.createAsyncTask(func() {
		e.viewParams.viewLineCount = viewLineCount
		e.viewParams.viewLineRuneWidth = viewLineRuneWidth
		e.frame.SetMaxRuneWidth(viewLineRuneWidth)
	})
	return e.send(e.ctx, msg)
}

// SetTextUnitPx sets text unit size in px.
// The text unit size means pixel region of drawn half character on UI.
// This setting affects calculation of image size in Editor.
func (e *Editor) SetTextUnitPx(textUnitPx image.Point) error {
	msg := e.createAsyncTask(func() {
		// create new loader instance to invalid internal cache.
		e.imgLoader = NewImageBytesLoader(
			DefaultCachedImageSize,
			textUnitPx.X,
			textUnitPx.Y,
			e.opt.ImageFetchType,
		)
	})
	return e.send(e.ctx, msg)
}

func (e *Editor) createCurrentParagraph(fixed bool) *pubdata.Paragraph {
	var lastP *text.Paragraph = nil
	for pp := e.frame.FirstParagraph(); pp != nil; pp = pp.Next(e.frame) {
		lastP = pp
	}
	if lastP == nil {
		panic("Paragraph is not found")
	}

	lines := make([]pubdata.Line, 0, lastP.LineCount(e.frame))
	for ll := lastP.FirstLine(e.frame); ll != nil; ll = ll.Next(e.frame) {
		boxes := make([]pubdata.Box, 0, 2)
		totalW := 0
		for bb := ll.FirstBox(e.frame); bb != nil; bb = bb.Next(e.frame) {
			newBB := e.createBox(bb)
			boxes = append(boxes, newBB)
			totalW += newBB.RuneWidth()
		}
		lines = append(lines, pubdata.Line{
			Boxes:     pubdata.NewBoxes(boxes),
			RuneWidth: totalW,
		})
	}
	alignment := pubdata.AlignmentString(attribute.Alignment(e.editor.GetAlignment()))
	return &pubdata.Paragraph{
		ID:        int(e.publishedCount % math.MaxInt32),
		Lines:     pubdata.NewLines(lines),
		Alignment: alignment,
		Fixed:     fixed,
	}
}

func (e *Editor) createBox(bb text.Box) pubdata.Box {
	bcommon := pubdata.BoxCommon{
		CommonRuneWidth:     bb.RuneWidth(e.frame),
		CommonLineCountHint: bb.LineCountHint(e.frame),
	}
	switch typed_bb := bb.(type) {
	case text.ImageBox:
		bcommon.CommonContentType = pubdata.ContentTypeImage
		return e.createImageBox(bcommon, typed_bb)
	case text.ButtonBox:
		bcommon.CommonContentType = pubdata.ContentTypeTextButton
		return &pubdata.TextButtonBox{
			BoxCommon: bcommon,
			BoxData: pubdata.TextButtonData{
				TextData: pubdata.TextData{
					Text:    typed_bb.Text(e.frame),
					FgColor: int(ColorRGBAToInt32RGB(typed_bb.FgColor())),
					BgColor: int(ColorRGBAToInt32RGB(text.ResetColor)),
				},
				Command: typed_bb.Command(),
			},
		}
	case text.TextBox:
		bcommon.CommonContentType = pubdata.ContentTypeText
		return &pubdata.TextBox{
			BoxCommon: bcommon,
			BoxData: pubdata.TextData{
				Text:    typed_bb.Text(e.frame),
				FgColor: int(ColorRGBAToInt32RGB(typed_bb.FgColor())),
				BgColor: int(ColorRGBAToInt32RGB(text.ResetColor)),
			},
		}
	case text.LineBox:
		bcommon.CommonContentType = pubdata.ContentTypeText
		return &pubdata.TextBox{
			BoxCommon: bcommon,
			BoxData: pubdata.TextData{
				Text:    typed_bb.Text(e.frame),
				FgColor: int(ColorRGBAToInt32RGB(typed_bb.FgColor())),
				BgColor: int(ColorRGBAToInt32RGB(text.ResetColor)),
			},
		}
	default:
		panic("unknown text.Box type")
	}
}

func (e *Editor) createImageBox(bcommon pubdata.BoxCommon, imgBox text.ImageBox) *pubdata.ImageBox {
	loadResult := e.imgLoader.LoadBytes(imgBox.SourceImage(), bcommon.RuneWidth(), bcommon.LineCountHint())
	// replace by resolved image size.
	bcommon.CommonRuneWidth = loadResult.TsSize.Width
	bcommon.CommonLineCountHint = loadResult.TsSize.Height
	// create image box
	return &pubdata.ImageBox{
		BoxCommon: bcommon,
		BoxData: pubdata.ImageData{
			Source:          imgBox.SourceImage(),
			WidthPx:         loadResult.PxSize.X,
			HeightPx:        loadResult.PxSize.Y,
			WidthTextScale:  bcommon.RuneWidth(),
			HeightTextScale: bcommon.LineCountHint(),
			Data:            loadResult.Bytes,
			DataFetchType:   loadResult.FetchType,
		}}
}

// ===== uiadapter.Printer interface APIs ======

// Sync flushes any pending output result, PrintXXX or ClearLine,
// at UI implementor. It can also use rate limitting for PrintXXX functions.
func (e *Editor) Sync() error {
	var taskErr error = nil
	msg := e.createSyncTask(func() {
		p := e.createCurrentParagraph(false)
		taskErr = e.callback.OnPublishTemporary(p)
	})
	if err := e.sendAndWait(e.ctx, msg); err != nil {
		return err
	}
	return taskErr
}

// Print text to screen.
// It should implement moving to next line by "\n".
func (e *Editor) Print(s string) error {
	msg := e.createAsyncTask(func() {
		if err := e.printInternal(s); err != nil {
			e.setAsyncErr(err)
			e.looper.Close()
		}
	})
	return e.send(e.ctx, msg)
}

func (e *Editor) printInternal(s string) error {
	for len(s) > 0 {
		si := strings.Index(s, "\n")
		if si == -1 {
			// write text full
			_, err := e.editor.WriteText(s)
			return err
		}
		// write text before first "\n"
		if _, err := e.editor.WriteText(s[:si]); err != nil {
			return err
		}
		// publish current paragraph ends by "\n"
		if err := e.publishParagraph(); err != nil {
			return err
		}
		// next text line exclude first "\n"
		s = s[si+1:]
	}
	return nil
}

func (e *Editor) publishParagraph() error {
	p := e.createCurrentParagraph(true)
	if err := e.callback.OnPublish(p); err != nil {
		return err
	}
	e.editor.DeleteLastParagraphs(1) // delete published content
	e.publishedCount++
	return nil
}

// Print label text to screen.
// It should not be separated in wrapping text.
func (e *Editor) PrintLabel(s string) error {
	msg := e.createAsyncTask(func() {
		if _, err := e.editor.WriteLabel(s); err != nil {
			e.setAsyncErr(err)
			e.looper.Close()
		}
	})
	return e.send(e.ctx, msg)
}

// Print Clickable button text. it shows caption on screen and emits command
// when it is selected. It is no separatable in wrapping text.
func (e *Editor) PrintButton(caption string, command string) error {
	msg := e.createAsyncTask(func() {
		if _, err := e.editor.WriteButton(caption, command); err != nil {
			e.setAsyncErr(err)
			e.looper.Close()
		}
	})
	return e.send(e.ctx, msg)
}

// Print Line using sym.
// given sym #, output line is: ############...
func (e *Editor) PrintLine(sym string) error {
	msg := e.createAsyncTask(func() {
		if err := e.printLineInternal(sym); err != nil {
			e.handleAsyncErr(err)
		}
	})
	return e.send(e.ctx, msg)
}

func (e *Editor) printLineInternal(sym string) error {
	// NOTE: Because e.editor.WriteLine appends new Paragraph after write text line.
	// currentParagraph is always empty and not publish the text line.
	// simulate the behavior at upper layer instead of use that API.

	// PrintLine should publish current content if have any.
	if e.editor.CurrentRuneWidth() > 0 {
		if err := e.publishParagraph(); err != nil {
			return err
		}
	}
	// Line always fills one line, publish it.
	nSym := e.viewParams.viewLineRuneWidth / width.StringWidth(sym)
	txtLine := strings.Repeat(sym, nSym)
	if _, err := e.editor.WriteText(txtLine); err != nil {
		return err
	}
	if err := e.publishParagraph(); err != nil {
		return err
	}
	return nil
}

// Print image from file path.
// Image is exceptional case, which may draw image region exceed over 1 line.
func (e *Editor) PrintImage(file string, widthInRW, heightInLC int) error {
	msg := e.createAsyncTask(func() {
		if err := e.editor.WriteImage(file, widthInRW, heightInLC); err != nil {
			e.setAsyncErr(err)
			e.looper.Close()
		}
	})
	return e.send(e.ctx, msg)
}

// Measure Image size in text scale, width in rune-width and height in line-count.
// This is useful when PrintImage will call with either widthInRW or heightInLC is zero,
// the drawn image size shall be auto determined but client want to know determined size
// before calling PrintImage.
func (e *Editor) MeasureImageSize(file string, widthInRW, heightInLC int) (retW, retH int, err error) {
	msg := e.createSyncTask(func() {
		ret := e.imgLoader.LoadBytes(file, widthInRW, heightInLC)
		retW = ret.TsSize.Width
		retH = ret.TsSize.Height
	})
	err = e.sendAndWait(e.ctx, msg)
	return
}

// Set and Get Color using 0xRRGGBB for 24bit color
func (e *Editor) SetColor(color uint32) error {
	msg := e.createAsyncTask(func() {
		e.editor.Color = Int32RGBToColorRGBA(int32(color))
	})
	return e.send(e.ctx, msg)
}

func (e *Editor) GetColor() (color uint32, err error) {
	msg := e.createSyncTask(func() {
		color = uint32(ColorRGBAToInt32RGB(e.editor.Color))
	})
	err = e.sendAndWait(e.ctx, msg)
	return
}

func (e *Editor) ResetColor() error {
	msg := e.createAsyncTask(func() {
		e.editor.Color = text.ResetColor
	})
	return e.send(e.ctx, msg)
}

// Set and Get Alignment
func (e *Editor) SetAlignment(align attribute.Alignment) error {
	msg := e.createAsyncTask(func() {
		e.editor.SetAlignment(text.Alignment(align)) // alignment is same constant value
	})
	return e.send(e.ctx, msg)
}

func (e *Editor) GetAlignment() (align attribute.Alignment, err error) {
	msg := e.createSyncTask(func() {
		align = attribute.Alignment(e.editor.GetAlignment()) // alignment is same constant value
	})
	err = e.sendAndWait(e.ctx, msg)
	return
}

// skip current lines to display none.
// TODO: is it needed?
func (e *Editor) NewPage() error {
	msg := e.createAsyncTask(func() {
		n := e.viewParams.viewLineCount
		for i := 0; i < n; i++ {
			if err := e.publishParagraph(); err != nil {
				e.setAsyncErr(err)
				e.looper.Close()
				return
			}
		}
	})
	return e.send(e.ctx, msg)
}

// Clear lines specified number.
func (e *Editor) ClearLine(nline int) error {
	msg := e.createAsyncTask(func() {
		if nline <= 0 {
			return
		}
		removeN := nline - 1
		if err := e.callback.OnRemove(removeN); err != nil {
			e.setAsyncErr(err)
			e.looper.Close()
			return
		}
		e.editor.DeleteLastParagraphs(1) // delete current editing line too
	})
	return e.send(e.ctx, msg)
}

// Clear all lines containing historys.
func (e *Editor) ClearLineAll() error {
	msg := e.createAsyncTask(func() {
		if err := e.callback.OnRemoveAll(); err != nil {
			e.handleAsyncErr(err)
			return
		}
		e.editor.DeleteLastParagraphs(1) // delete current editing line too
	})
	return e.send(e.ctx, msg)
}

// rune width to fill the window width.
func (e *Editor) WindowRuneWidth() (int, error) {
	var result int = 0
	msg := e.createSyncTask(func() {
		result = e.viewParams.viewLineRuneWidth
	})
	return result, e.sendAndWait(e.ctx, msg)
}

// line count to fill the window height.
func (e *Editor) WindowLineCount() (int, error) {
	var result int = 0
	msg := e.createSyncTask(func() {
		result = e.viewParams.viewLineCount
	})
	return result, e.sendAndWait(e.ctx, msg)
}

// current rune width in the editting line.
func (e *Editor) CurrentRuneWidth() (int, error) {
	var result int = 0
	msg := e.createSyncTask(func() {
		result = e.editor.CurrentRuneWidth()
	})
	return result, e.sendAndWait(e.ctx, msg)
}

// line count as it increases at outputting new line.
func (e *Editor) LineCount() (int, error) {
	var result int = 0
	msg := e.createSyncTask(func() {
		// NOTE: new line count is not changed at text.Editor, since
		// new line is handled by this Editor internally.
		// Use own publish count instead of text.Editor's one.
		result = int(e.publishedCount % math.MaxInt32)
		//result = e.editor.NewLineCount()
	})
	return result, e.sendAndWait(e.ctx, msg)
}
