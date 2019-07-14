package erago

import (
	attr "github.com/mzki/erago/attribute"
	"github.com/mzki/erago/scene"
	"github.com/mzki/erago/util/errutil"
	"github.com/mzki/erago/util/log"
)

// This files defines game scenes which does system level processure.

const (
	sceneNameBooting = "__booting"
	// TODO save/load scene is here?

	replaceTextFileName = "__builtin_replace.lua"
	replaceTextDataKey  = "ERA_BUILTIN_SCENE_REPLACE_TEXT"

	replaceTextLoadingMessage   = "LoadingMessage"
	replaceTextNewGame          = "NewGame"
	replaceTextLoadGame         = "LoadGame"
	replaceTextQuitGame         = "QuitGame"
	replaceTextReturnMenu       = "ReturnMenu"
	replaceTextMoneyFormat      = "MoneyFormat"
	replaceTextSelectSaveData   = "SelectSaveData"
	replaceTextSelectLoadData   = "SelectLoadData"
	replaceTextConfirmOverwrite = "ConfirmOverwirte"

	defaultLoadingMessage = "Now Loading..."
)

func (game *Game) sceneBooting() (string, error) {
	var replaceData map[string]string
	replaceData, err := game.ipr.LoadDataOnSandbox(game.ipr.PathOf(replaceTextFileName), replaceTextDataKey)
	if err != nil {
		log.Debugf("%s: Failed to load replace text from %s. No use replacement.", sceneNameBooting, replaceTextFileName)
		replaceData = map[string]string{}
	}

	// require parse format
	{
		moneyFormat := replaceData[replaceTextMoneyFormat]
		if parsed, err := scene.ParseMoneyFormat(moneyFormat); err != nil {
			log.Debugf("%s: Failed to parse money format for replacement text, %s", sceneNameBooting, moneyFormat)
			replaceData[replaceTextMoneyFormat] = "" // overwrite by the no replacement.
		} else {
			replaceData[replaceTextMoneyFormat] = parsed
		}
	}

	var replaceText scene.ConfigReplaceText
	replaceText.LoadingMessage = replaceData[replaceTextLoadingMessage]
	replaceText.NewGame = replaceData[replaceTextNewGame]
	replaceText.LoadGame = replaceData[replaceTextLoadGame]
	replaceText.QuitGame = replaceData[replaceTextQuitGame]
	replaceText.ReturnMenu = replaceData[replaceTextReturnMenu]
	replaceText.MoneyFormat = replaceData[replaceTextMoneyFormat]
	replaceText.SelectSaveData = replaceData[replaceTextSelectSaveData]
	replaceText.SelectLoadData = replaceData[replaceTextSelectLoadData]
	replaceText.ConfirmOverwrite = replaceData[replaceTextSelectLoadData]

	if err := replaceText.Validate(); err != nil {
		log.Debugf("%s: Invalid replace text: %v", sceneNameBooting, err)
		replaceText = scene.ConfigReplaceText{}
	}

	if err := game.scene.SetReplaceText(replaceText); err != nil {
		log.Debugf("%s: Failed to set replace text: %v", sceneNameBooting, err)
	}

	var LoadingMessage string = scene.DefaultOrString(defaultLoadingMessage, replaceText.LoadingMessage)

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
	err = game.ipr.LoadSystem()

	// next scene is title scene
	return scene.SceneNameTitle, err
}
