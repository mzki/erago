package model

import (
	"context"
	"reflect"
	"testing"

	"github.com/mzki/erago/uiadapter"
	"github.com/mzki/erago/view/exp/text/pubdata"
	"github.com/mzki/erago/view/exp/text/publisher"
)

type pseudoPublishedResult struct {
	List []*pubdata.Paragraph
}

func newPsedoPublishedResult() *pseudoPublishedResult {
	return &pseudoPublishedResult{
		List: make([]*pubdata.Paragraph, 0, 8),
	}
}

func (pres *pseudoPublishedResult) OnPublsh(p *pubdata.Paragraph) error {
	pres.List = append(pres.List, p)
	return nil
}

func (pres *pseudoPublishedResult) OnPublishTemporary(p *pubdata.Paragraph) error {
	// replacement will happen only in OnPublishTemporary
	for i, pp := range pres.List {
		if pp.Id == p.Id {
			pres.List[i] = p // to replace
			return nil
		}
	}
	pres.List = append(pres.List, p)
	return nil
}

func (pres *pseudoPublishedResult) OnRemove(n int) error {
	if l := pres.List; len(l) == 0 {
		return nil
	} else {
		// case of len(l) > 0
		last := l[len(l)-1]
		if !last.Fixed {
			makeEmptyParagraph(last)
			l = l[:len(l)-1] // to exclude remove operation
		}
		nRemove := n
		if nRemove > len(l) {
			nRemove = len(l)
		}

		l = l[:len(l)-nRemove]
		if !last.Fixed {
			l = append(l, last)
		}
		pres.List = l
	}
	return nil
}

func (pres *pseudoPublishedResult) OnRemoveAll() error {
	pres.List = pres.List[:0]
	return nil
}

func (pres *pseudoPublishedResult) OnSync() error {
	// do nothing
	return nil
}

type pseudoUiResult struct {
	Result       *pseudoPublishedResult
	InputRequest uiadapter.InputRequestType
}

func newPsedoUiResult() *pseudoUiResult {
	return &pseudoUiResult{
		Result:       newPsedoPublishedResult(),
		InputRequest: uiadapter.InputRequestNone,
	}
}

func (pui *pseudoUiResult) OnPublishBytes(bs []byte) error {
	plist := &pubdata.ParagraphList{}
	if err := plist.UnmarshalVT(bs); err != nil {
		return err
	}
	for _, p := range plist.Paragraphs {
		if p.Fixed {
			_ = pui.Result.OnPublsh(p) // never return error
		} else {
			_ = pui.Result.OnPublishTemporary(p) // never return error
		}
	}
	return nil
}

func (pui *pseudoUiResult) OnPublishBytesTemporary(bs []byte) error {
	// this will not used. but delegates to OnPublishBytes for fail safe.
	return pui.OnPublishBytes(bs)
}

func (pui *pseudoUiResult) OnRemove(n int) error {
	return pui.Result.OnRemove(n)
}

func (pui *pseudoUiResult) OnRemoveAll() error {
	return pui.Result.OnRemoveAll()
}

func (pui *pseudoUiResult) OnDebugTimestamp(int64, string, int64) error {
	return nil // do nothing
}

func (pui *pseudoUiResult) OnInputRequested() {
	pui.InputRequest = uiadapter.InputRequestInput
}

func (pui *pseudoUiResult) OnCommandRequested() {
	pui.InputRequest = uiadapter.InputRequestCommand
}

func (pui *pseudoUiResult) OnInputRequestClosed() {
	pui.InputRequest = uiadapter.InputRequestNone
}

func newResultPair(ctx context.Context) (*pseudoPublishedResult, *pseudoUiResult, *publisher.Editor, context.CancelFunc) {
	var deferredFuncs = make([]func(), 0, 2)
	var cancelFunc = func() {
		// reverse order to call deferredFuncs as same as the order of go's builtin defer.
		for i := len(deferredFuncs) - 1; i >= 0; i-- {
			deferredFuncs[i]()
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	deferredFuncs = append(deferredFuncs, cancel)

	pResult := newPsedoPublishedResult()
	pendingEvents := newPendingEvents()
	pUiResult := newPsedoUiResult()
	binEncode := newParagraphListBinaryEncodeFunc(MessageByteEncodingProtobuf)

	editor := publisher.NewEditor(ctx, publisher.EditorOptions{
		ImageFetchType: publisher.ImageFetchNone,
		ImageCacheSize: 16,
	})
	deferredFuncs = append(deferredFuncs, func() { editor.Close() })

	editor.SetCallback(&publisher.CallbackDefault{
		OnPublishFunc: func(p *pubdata.Paragraph) error {
			pendingEvents.appendAddParagraph(p)
			return pResult.OnPublsh(p)
		},
		OnPublishTemporaryFunc: func(p *pubdata.Paragraph) error {
			pendingEvents.appendAddParagraph(p)
			return pResult.OnPublishTemporary(p)
		},
		OnRemoveFunc: func(nParagraph int) error {
			pendingEvents.appendRemoveParagraph(nParagraph)
			return pResult.OnRemove(nParagraph)
		},
		OnRemoveAllFunc: func() error {
			pendingEvents.appendRemoveParagraphAll()
			return pResult.OnRemoveAll()
		},
		OnSyncFunc: func() error {
			err := pendingEvents.publish(pUiResult, binEncode)
			if err != nil {
				return err
			}
			// do nothing on pResult
			return nil
		},
	})

	return pResult, pUiResult, editor, cancelFunc
}

func testHelperCompareParagraphList(t *testing.T, pResult *pseudoPublishedResult, pUiResult *pseudoUiResult) {
	t.Helper()

	if len(pResult.List) != len(pUiResult.Result.List) {
		// This is fatal case actually, but just reporting error to check each element at following check.
		t.Errorf("different result length of paragraph list, expect: %v, got: %v", len(pResult.List), len(pUiResult.Result.List))
	}

	nMin := len(pResult.List)
	if uiLen := len(pUiResult.Result.List); nMin > uiLen {
		nMin = uiLen
	}
	for i := 0; i < nMin; i++ {
		p := pResult.List[i]
		uiP := pUiResult.Result.List[i]
		if !reflect.DeepEqual(p, uiP) {
			t.Errorf("different content at %v-th paragraph, expect: %v, got: %v", i, p, uiP)
		}
	}
}

func TestSameResultBetweenWithAndWithoutCallbackThrottler(t *testing.T) {
	fatalIfErr := func(t *testing.T, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("Fatal error: %v", err)
		}
	}
	t.Run("paragraph only, sync once", func(t *testing.T) {
		t.Parallel()
		pResult, pUiResult, editor, cancel := newResultPair(context.Background())
		defer cancel()

		editor.Print("abc\n")
		editor.PrintLabel("lable")
		editor.PrintLine("=")
		editor.Print("text\n")
		editor.Print("unfixed-text")
		fatalIfErr(t, editor.Sync())
		testHelperCompareParagraphList(t, pResult, pUiResult)
	})

	t.Run("paragraph only, sync twice", func(t *testing.T) {
		t.Parallel()
		pResult, pUiResult, editor, cancel := newResultPair(context.Background())
		defer cancel()

		editor.Print("abc\n")
		editor.PrintLabel("lable")
		editor.PrintLine("=")
		editor.Print("text\n")
		editor.Print("unfixed-text")
		fatalIfErr(t, editor.Sync())

		editor.Print("fixed text\n")
		editor.PrintButton("lable", "command")
		editor.PrintSpace(8)
		editor.Print("text\n")
		fatalIfErr(t, editor.Sync())

		testHelperCompareParagraphList(t, pResult, pUiResult)
	})

	t.Run("paragraph(unfixed) -> remove (paragprah > remove), sync once", func(t *testing.T) {
		t.Parallel()
		pResult, pUiResult, editor, cancel := newResultPair(context.Background())
		defer cancel()

		editor.Print("abc\n")
		editor.PrintLabel("lable")
		editor.PrintLine("=")
		editor.Print("text\n")
		editor.Print("unfixed-text")
		editor.ClearLine(2)
		fatalIfErr(t, editor.Sync())

		testHelperCompareParagraphList(t, pResult, pUiResult)
	})

	t.Run("paragraph(unfixed) -> remove (paragprah > remove), sync twice", func(t *testing.T) {
		t.Parallel()
		pResult, pUiResult, editor, cancel := newResultPair(context.Background())
		defer cancel()

		editor.Print("abc\n")
		editor.PrintLabel("lable")
		editor.PrintLine("=")
		editor.Print("text\n")
		editor.Print("unfixed-text")
		fatalIfErr(t, editor.Sync())

		editor.ClearLine(4)
		fatalIfErr(t, editor.Sync())

		testHelperCompareParagraphList(t, pResult, pUiResult)
	})

	t.Run("paragraph(unfixed) -> remove (paragprah < remove), sync twice", func(t *testing.T) {
		t.Parallel()
		pResult, pUiResult, editor, cancel := newResultPair(context.Background())
		defer cancel()

		editor.Print("abc\n")
		editor.PrintLabel("lable")
		editor.PrintLine("=")
		editor.Print("text\n")
		editor.Print("unfixed-text")
		fatalIfErr(t, editor.Sync())

		editor.ClearLine(8)
		fatalIfErr(t, editor.Sync())

		testHelperCompareParagraphList(t, pResult, pUiResult)
	})

	t.Run("paragraph(fixed) -> paragraph(fixed) -> remove (paragprah < remove), sync twice", func(t *testing.T) {
		t.Parallel()
		pResult, pUiResult, editor, cancel := newResultPair(context.Background())
		defer cancel()

		editor.Print("abc\n")
		editor.PrintLabel("lable")
		editor.PrintLine("=")
		editor.Print("text\n")
		fatalIfErr(t, editor.Sync())

		editor.Print("efg\nhig\n")
		editor.ClearLine(10)
		fatalIfErr(t, editor.Sync())

		testHelperCompareParagraphList(t, pResult, pUiResult)
	})

	t.Run("paragraph(fixed) -> paragraph(unfixed) -> remove (paragprah < remove), sync twice", func(t *testing.T) {
		t.Parallel()
		pResult, pUiResult, editor, cancel := newResultPair(context.Background())
		defer cancel()

		editor.Print("abc\n")
		editor.PrintLabel("lable")
		editor.PrintLine("=")
		editor.Print("text\n")
		fatalIfErr(t, editor.Sync())

		editor.Print("unfixed")
		editor.ClearLine(8)
		fatalIfErr(t, editor.Sync())

		testHelperCompareParagraphList(t, pResult, pUiResult)
	})

	t.Run("paragraph(fixed) -> remove (paragprah > remove) -> paragraph(fixed) -> remove (paragprah > remove), sync twice", func(t *testing.T) {
		t.Parallel()
		pResult, pUiResult, editor, cancel := newResultPair(context.Background())
		defer cancel()

		editor.Print("abc\n")
		editor.PrintLabel("lable")
		editor.PrintLine("=")
		editor.Print("text\n")
		fatalIfErr(t, editor.Sync())

		editor.ClearLine(2)
		editor.Print("fixed\nfixed2\n")
		editor.ClearLine(2)
		fatalIfErr(t, editor.Sync())

		testHelperCompareParagraphList(t, pResult, pUiResult)
	})

	t.Run("paragraph(fixed) -> remove all -> paragraph(unfixed), sync twice", func(t *testing.T) {
		t.Parallel()
		pResult, pUiResult, editor, cancel := newResultPair(context.Background())
		defer cancel()

		editor.Print("abc\n")
		editor.PrintLabel("lable")
		editor.PrintLine("=")
		editor.Print("text\n")
		fatalIfErr(t, editor.Sync())

		editor.ClearLineAll()
		editor.Print("fixed\nunfixed")
		fatalIfErr(t, editor.Sync())

		testHelperCompareParagraphList(t, pResult, pUiResult)
	})

}
