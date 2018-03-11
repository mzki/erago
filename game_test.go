package erago

import (
	"testing"
	"time"

	"local/erago/scene"
	"local/erago/stub"
	"local/erago/uiadapter/event/input"
)

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

	g.scene.RegisterScene(scene.SceneNameTitle, scene.NextFunc(func() (string, error) {
		g.uiAdapter.Print("This is test scene first.\n")
		return "unknown", nil
	}))

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
