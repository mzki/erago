package erago

import (
	"fmt"
	"reflect"

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

	defaultLoadingMessage = "Now Loading..."
)

func (game *Game) sceneBooting() (string, error) {
	var replaceData map[string]string
	replaceData, err := game.ipr.LoadDataOnSandbox(game.ipr.PathOf(replaceTextFileName), replaceTextDataKey)
	if err != nil {
		log.Debugf("%s: Failed to load replace text from %s. No use replacement.", sceneNameBooting, replaceTextFileName)
		replaceData = map[string]string{}
	}

	var replaceText scene.ConfigReplaceText

	// fill replaceText by user defined data
	if err := fillStructByMap(&replaceText, replaceData); err != nil {
		log.Debugf("%s: Invalid replace format: %v", sceneNameBooting, err)
		return "", err
	}

	// require parse format for specific fields
	{
		moneyFormat := replaceText.MoneyFormat
		if parsed, err := scene.ParseMoneyFormat(moneyFormat); err != nil {
			log.Debugf("%s: Failed to parse money format for replacement text, %s", sceneNameBooting, moneyFormat)
			replaceText.MoneyFormat = "" // overwrite by the no replacement.
		} else {
			replaceText.MoneyFormat = parsed
		}
	}

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

func fillStructByMap(dst interface{}, src map[string]string) error {
	structValue := reflect.ValueOf(dst).Elem()
	for k, v := range src {
		field := structValue.FieldByName(k)
		if !field.IsValid() {
			continue // no such field, ignore
		}
		if !field.CanSet() {
			continue // read only field, ignore
		}

		val := reflect.ValueOf(interface{}(v))
		if field.Type() != val.Type() {
			return fmt.Errorf("map value type for key %s does not match struct field type", k)
		}

		field.Set(val)
	}
	return nil
}
