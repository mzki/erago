package erago

import (
	attr "github.com/mzki/erago/attribute"
	"github.com/mzki/erago/scene"
	"github.com/mzki/erago/util/errutil"
)

// This files defines game scenes which does system level processure.

const (
	sceneNameBooting = "__booting"
	// TODO save/load scene is here?
)

func (game *Game) sceneBooting() (string, error) {
	var LoadingMessage = game.state.CSV.GameReplace.LoadingMessage

	ui := game.uiAdapter
	merr := errutil.NewMultiError()
	merr.Add(ui.SetAlignment(attr.AlignmentRight))
	merr.Add(ui.NewPage())
	merr.Add(ui.PrintL(LoadingMessage))
	merr.Add(ui.SetAlignment(attr.AlignmentLeft))

	if err := merr.Err(); err != nil {
		return "", err
	}

	// extract user scripts and register it.
	err := game.ipr.LoadSystem()

	// next scene is title scene
	return scene.SceneNameTitle, err
}