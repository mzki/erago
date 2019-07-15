package scene

import (
	"context"
	"fmt"
	"time"

	"github.com/mzki/erago/state"
	"github.com/mzki/erago/util/log"
)

const (
	// Scene LoadGame and SaveGame are not registered in default scene transition.
	// These are used by call SceneManager.RunXXXScene() from external.
	SceneNameLoadGame = "loadgame"
	SceneNameSaveGame = "savegame"
)

// * LOAD GAME SCENE
type loadGameScene struct {
	sceneCommon
}

func newLoadGameScene(sf *sceneFields) *loadGameScene {
	return &loadGameScene{newSceneCommon(SceneNameLoadGame, sf)}
}

func (lg *loadGameScene) Name() string { return SceneNameLoadGame }

func (lg *loadGameScene) Next() (Scene, error) {
	if next, err := lg.atStart(); next != nil || err != nil {
		return next, err
	}
	return loadGameSceneProcess(lg.sceneFields)
}

// return next scene `loaded` if load success.
func loadGameSceneProcess(sf *sceneFields) (Scene, error) {
	game := sf.IO()

	game.PrintL(DefaultOrString("Select Load Data >>", sf.ReplaceText().SelectLoadData))
	printSaveListsScene(sf)

	for {
		input, err := game.CommandNumber()
		if err != nil {
			return nil, err
		}
		switch {
		case 100 == input:
			return nil, nil

		case 0 <= input && input < 20 || input == 99:
			gstate := sf.State()
			if gstate.FileExists(input) {
				if err := gstate.LoadSystem(input); err != nil {
					return nil, err
				}
				return sf.Scenes().GetScene(SceneNameLoadEnd)
			}
		} // .. switch
	}
}

// * SAVE GAME SCENE
type saveGameScene struct {
	sceneCommon
}

func newSaveGameScene(sf *sceneFields) *saveGameScene {
	return &saveGameScene{newSceneCommon(SceneNameSaveGame, sf)}
}

func (sg *saveGameScene) Name() string { return SceneNameSaveGame }

func (sg *saveGameScene) Next() (Scene, error) {
	if next, err := sg.atStart(); next != nil || err != nil {
		return next, err
	}

	game := sg.IO()
	// TODO: current layout is resereved and revert after save/load?
	// game.ReserveLayout()
	// defer game.SetPreviousLayout()
	// game.SetSingleLayout()

	for {
		game.PrintL(DefaultOrString("Select Save Destination >>", sg.ReplaceText().SelectSaveData))
		printSaveListsScene(sg.sceneFields)

		input, err := game.CommandNumber()
		if err != nil {
			return nil, err
		}
		switch {
		case 100 == input:
			goto END_SAVE_GAME

		case 0 <= input && input < 20:
			gstate := sg.State()
			if gstate.FileExists(input) {
				game.PrintL(DefaultOrString("Overwrite?", sg.ReplaceText().ConfirmOverwrite))
				game.PrintC("[0] Yes", 10)
				game.PrintC("[1] No", 10)
				game.PrintL("")
				if yesno, err := game.CommandNumberSelect(context.Background(), 0, 1); err != nil {
					return nil, err
				} else if yesno == 1 {
					game.PrintL("")
					break // switch
				}
			}

			if err := saveGameSceneProcess(input, sg.sceneFields); err != nil {
				return nil, err
			}
			goto END_SAVE_GAME
		}
	}

END_SAVE_GAME:
	next := sg.Scenes().Prev()
	return next, nil
}

// +scene: savegame
// savegame シーンは特殊なシーンです。
// 他の組み込みシーンとは違い、直接遷移することはできません
const (
	// +callback: {{.Name}}()
	// データが保存される直前に呼び出されます。
	// ここで、セーブデータのコメントを書き換えることができます。
	// デフォルトでは、 "2006/01/02 15:04:05" の形式で、現在時刻がコメントとして使用されます
	ScrSaveGameEventBeforeSave = "savegame_event_before_save"
)

func saveGameSceneProcess(No int, sf *sceneFields) error {
	sf.State().SaveComment = time.Now().Format("2006/01/02 15:04:05")
	if err := sf.Script().maybeCall(ScrSaveGameEventBeforeSave); err != nil {
		return err
	}
	return sf.State().SaveSystem(No)
}

// * LOAD END SCENE
type loadEndScene struct {
	sceneCommon
}

func newLoadEndScene(sf *sceneFields) *loadEndScene {
	return &loadEndScene{newSceneCommon(SceneNameLoadEnd, sf)}
}

func (ld *loadEndScene) Name() string { return SceneNameLoadEnd }

// +scene: loadend
const (
// ScrSceneLoadEnd = "scene_loadend"
// ScrEventLoadEnd = "event_loadend"
)

func (ld *loadEndScene) Next() (Scene, error) {
	if next, err := ld.atStart(); next != nil || err != nil {
		return next, err
	}

	if ss := ld.Scenes(); ss.HasNext() {
		return ss.Next(), nil
	} else {
		return ss.GetScene(SceneNameBase)
	}
}

func printSaveListsScene(sf *sceneFields) {
	printSaveLists(sf)
	sf.IO().PrintLine(DefaultLineSymbol)
	sf.IO().PrintL("[100] " + DefaultOrString("Back", sf.ReplaceText().ReturnMenu))
}

const autoSaveNumber = 99

func printSaveLists(sf *sceneFields) {
	buildSaveTitle := func(list []*state.MetaData, i, no int) string {
		save_title := fmt.Sprintf("[%2d] ", no)

		if header := list[i]; header == nil {
			save_title += "----"
		} else {
			save_title += header.Title
		}
		return save_title
	}

	list := buildHeaderLists(sf.State())
	for i := 0; i < len(list)-1; i++ {
		sf.IO().PrintL(buildSaveTitle(list, i, i))
	}
	// auto save number
	sf.IO().PrintL(buildSaveTitle(list, len(list)-1, autoSaveNumber))
}

func buildHeaderLists(gstate *state.GameState) []*state.MetaData {
	list := make([]*state.MetaData, 21)

	for i := 0; i < 20; i++ {
		if gstate.FileExists(i) {
			if header, err := gstate.LoadHeader(i); err != nil {
				log.Debug("buildHeaderLists: ", header, err)
			} else {
				list[i] = header
			}
		}
	}
	// auto save number
	if header, err := gstate.LoadHeader(autoSaveNumber); err != nil {
		log.Debug("load autosave header: ", header, err)
	} else {
		list[20] = header
	}
	return list
}
