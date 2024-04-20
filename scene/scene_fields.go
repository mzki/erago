package scene

import (
	"fmt"

	"github.com/mzki/erago/state"
)

// common members of scene.
type sceneFields struct {
	// packed struct
	io          IOController
	callbacker  callBacker
	state       *state.GameState
	scenes      *sceneHolder
	conf        Config
	replaceText ConfigReplaceText
}

// Get Field Methods.
func (sf sceneFields) IO() IOController               { return sf.io }
func (sf sceneFields) Script() callBacker             { return sf.callbacker }
func (sf sceneFields) Scenes() *sceneHolder           { return sf.scenes }
func (sf sceneFields) State() *state.GameState        { return sf.state }
func (sf sceneFields) Config() Config                 { return sf.conf }
func (sf sceneFields) ReplaceText() ConfigReplaceText { return sf.replaceText }

// sceneHolder holds secne instances and next and prev scene.
type sceneHolder struct {
	prev Scene
	next Scene

	scenes map[string]Scene
}

const (
	// use for get or set scene name.
	SceneNameTitle    = "title"
	SceneNameNewGame  = "newgame"
	SceneNameAutosave = "autosave"
	SceneNameBase     = "base"
	SceneNameShop     = "shop"
	SceneNameTrain    = "train"
	SceneNameAblUp    = "ablup"
	SceneNameTrainEnd = "trainend"
	SceneNameTurnEnd  = "turnend"
	SceneNameLoadEnd  = "loadend"
)

func newSceneHolder(sf *sceneFields) *sceneHolder {
	shr := &sceneHolder{
		prev: nil,
		next: nil,
	}

	shr.scenes = map[string]Scene{
		SceneNameTitle:    newTitleScene(sf),
		SceneNameNewGame:  newNewGameScene(sf),
		SceneNameAutosave: newAutosaveScene(sf),
		SceneNameBase:     newBaseScene(sf),
		SceneNameShop:     newShopScene(sf),
		SceneNameTrain:    newTrainScene(sf),
		SceneNameAblUp:    newAblUpScene(sf),
		SceneNameTrainEnd: newTrainEndScene(sf),
		SceneNameTurnEnd:  newTurnEndScene(sf),
		SceneNameLoadEnd:  newLoadEndScene(sf),
	}
	return shr
}

func (sh sceneHolder) Next() Scene { return sh.next }
func (sh sceneHolder) Prev() Scene { return sh.prev }

func (sh sceneHolder) HasNext() bool { return sh.next != nil }
func (sh sceneHolder) HasPrev() bool { return sh.prev != nil }

func (sh *sceneHolder) SetNext(s Scene) Scene {
	sh.next = s
	return sh.next
}

func (sh *sceneHolder) SetNextByName(name string) error {
	scene, err := sh.GetScene(name)
	if err != nil {
		return err
	}
	sh.SetNext(scene)
	return nil
}

func (sh *sceneHolder) SetPrev(s Scene) Scene {
	sh.prev = s
	return sh.prev
}

// GetScene returns Scene specifyed by name.
// If name is not registered yet, it returns nil scene and error including ErrorSceneNameNotRegistered.
func (sh sceneHolder) GetScene(name string) (Scene, error) {
	s, ok := sh.scenes[name]
	if ok {
		return s, nil
	}
	return nil, fmt.Errorf(`scene "%v" is not found: %w`, name, ErrorSceneNameNotRegistered)
}

// register scene to add new flow for the scene transition.
func (sh sceneHolder) registerScene(s Scene) {
	sh.scenes[s.Name()] = s
}

// unregister scene to remove new flow from the scene transition.
func (sh sceneHolder) unRegisterScene(s Scene) {
	delete(sh.scenes, s.Name())
}
