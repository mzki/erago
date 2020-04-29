package erago

import (
	"testing"
	"time"

	"github.com/mzki/erago/scene"
	"github.com/mzki/erago/stub"
	"github.com/mzki/erago/uiadapter/event/input"
	"github.com/mzki/erago/util/log"
)

func init() {
	log.SetLevel(log.InfoLevel)
}

func TestGameMain(t *testing.T) {
	g := NewGame()
	if err := g.Init(stub.NewGameUIStub(), NewConfig("./stub")); err != nil {
		t.Fatal(err)
	}
	defer g.Quit()

	// send quit signal to prevent never end of running game.
	go func() {
		time.Sleep(1 * time.Second)
		g.Send(input.NewEventQuit())
	}()

	g.scene.RegisterSceneFunc(scene.SceneNameTitle, func() (string, error) {
		g.uiAdapter.Print("This is test scene first.\n")
		return "unknown", nil
	})

	if err := g.Main(); err == nil {
		t.Error("must be error but no erorr")
	} else {
		t.Log(err)
	}
}

func TestMainQuitExternally(t *testing.T) {
	g := NewGame()
	if err := g.Init(stub.NewGameUIStub(), NewConfig("./stub")); err != nil {
		t.Fatal(err)
	}
	defer g.Quit()

	// send quit signal to prevent never end of running game.
	go func() {
		time.Sleep(1 * time.Second)
		g.Quit()
	}()

	if err := g.Main(); err != nil {
		t.Errorf("must quit correctly, but error: %v", err)
	}
}
