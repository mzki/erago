package mobile

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"testing"
	"time"
)

type appContextStub struct{}

func (appContextStub) NotifyQuit(err error) {
	fmt.Println(err)
}
func (appContextStub) NotifyPaint() {
	b := LockBuffer()
	defer UnlockBuffer(b)
}
func (appContextStub) NotifyCommandRequest(cmds *CmdSlice) {
	fmt.Println(cmds)
}
func (appContextStub) NotifyCommandRequestClose() {
}

var screenSize = image.Point{256, 256}

func TestApp(t *testing.T) {
	OnStart("../stub/", appContextStub{})
	defer OnStop()

	OnMeasure(screenSize.X, screenSize.Y, 72)
	time.Sleep(1 * time.Second) // wait for initializing app.

	ch := make(chan struct{})
	func() {
		defer close(ch)
		time.Sleep(100 * time.Millisecond)
		OnCommandSelected("test")
	}()
	<-ch

	rgba := image.NewRGBA(image.Rectangle{Max: screenSize})
	bs := LockBuffer()
	copy(rgba.Pix, bs)
	UnlockBuffer(bs)

	if err := writePng("testimage/_app_test.png", rgba); err != nil {
		t.Error(err)
	}
}

func writePng(file string, rgba image.Image) error {
	fp, err := os.Create(file)
	if err != nil {
		return err
	}
	defer fp.Close()

	return png.Encode(fp, rgba)
}
