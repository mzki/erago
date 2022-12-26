package scene

import (
	"context"
	"fmt"

	"github.com/mzki/erago/state"
	"github.com/mzki/erago/state/csv"
	"github.com/mzki/erago/util/log"
)

// SceneManager is entry point of the scene flow transition.
//
// Example:
//
//	  sm := NewSceneManager(...)
//	  defer sm.Free()
//		 // do something
type SceneManager struct {
	sf *sceneFields

	currentScene Scene
}

func NewSceneManager(game IOController, scr Scripter, state *state.GameState, config Config) *SceneManager {
	sf := &sceneFields{
		callbacker:  callBacker{&loggedScripter{scr}, game},
		io:          game,
		conf:        config,
		state:       state,
		replaceText: ConfigReplaceText{},
	}

	sh := newSceneHolder(sf)
	sf.scenes = sh // NOTE: cross referene

	sm := &SceneManager{
		sf: sf,
	}
	return sm
}

// Free referene cycle. It must be called
// at end use of SceneManager for GC.
func (sm *SceneManager) Free() {
	sm.sf.io = nil
	sm.sf.callbacker = callBacker{}
	sm.sf.state = nil
	sm.sf.scenes = nil
}

// run scene transitions starting from start_scene.
// it blocks until done, you can use go func() to avoid blocking main thread.
func (sm *SceneManager) Run(ctx context.Context, start_scene string) (err error) {
	sceneHolder := sm.sf.Scenes()
	sm.currentScene, err = sceneHolder.GetScene(start_scene)
	if err != nil {
		return
	}

	for {
		// check cancelaration.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		log.Debug("SceneManager.Run(): starting scene ", sm.currentScene.Name())

		next, err := sm.currentScene.Next()

		switch err {
		case nil:
			// no error, do nothing.
		case ErrorSceneNext:
			// indicates force moving to next scene.
			next = sm.sf.Scenes().Next()
			if next == nil {
				return fmt.Errorf("SceneManager.Run(): got going to next scene, but next scene does not set")
			}
		case ErrorQuit:
			// indicates force quit or normal termination.
			return nil
		default:
			log.Debugf("SceneManager.Run(): %v in %v", err, sm.currentScene.Name())
			return err // error context, example is uiadpter.ErrorPipelineClosed, is remained.
		}

		if next == nil {
			return fmt.Errorf("SceneManager.Run(): scene %v returns nil as next scene", sm.currentScene.Name())
		}

		sceneHolder.SetPrev(sm.currentScene)
		sm.currentScene = next
		sceneHolder.SetNext(nil)
	}
	// panic("never reached")
}

// Run flow of scene: SaveGame.
func (sm SceneManager) DoSaveGameScene() error {
	_, err := newSaveGameScene(sm.sf).Next()
	return err
}

// Run flow of scene: LoadGame.
// It sets next scene `loadend` and return ErrorSceneNext if load success.
func (sm SceneManager) DoLoadGameScene() error {
	next, err := newLoadGameScene(sm.sf).Next()
	if err != nil {
		return err
	}
	if next != nil {
		sm.sf.scenes.SetNext(next)
		return ErrorSceneNext
	}
	return nil
}

// do train flow using commands which contains train No.
func (sm SceneManager) DoTrainsScene(commands []int64) error {
	if current := sm.currentScene.Name(); current != SceneNameTrain {
		return fmt.Errorf("do train can not be called from scene %s", current)
	}

	scene_train := sm.sf.Scenes().scenes[SceneNameTrain].(*trainScene)
	train_names := sm.sf.state.CSV.MustConst(csv.BuiltinTrainName).Names

	for _, cmd_no := range commands {
		if err := scene_train.CheckTrainCommand(int(cmd_no), train_names[cmd_no]); err != nil {
			return err
		}
		if err := scene_train.DoTrain(cmd_no); err != nil {
			return err
		}
		// TODO: stopping doTrain is not able for the character state.
		// NOW: stop by setted next scene.
		if sm.sf.Scenes().HasNext() {
			break
		}
	}
	return nil
}

// Set ConfigReplaceText to replace text in the builtin scene flow.
// It is concurrency unsafe.
func (sm *SceneManager) SetReplaceText(config ConfigReplaceText) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("scene: invalid replace text. %v", err)
	}
	sm.sf.replaceText = config
	return nil
}

// Set Next Scene using sence name, if scene name is
// not found return error.
func (sm SceneManager) SetNextSceneByName(scene_name string) error {
	return sm.sf.scenes.SetNextByName(scene_name)
}

// register user-defined scene into scene trainsition
func (sm SceneManager) RegisterScene(s Scene) {
	sm.sf.scenes.registerScene(s)
}

// register user-defined scene flow into scene trainsition
// using scene name and next function.
// next function is converted scene.NextFunc internally.
func (sm SceneManager) RegisterSceneFunc(name string, next_func func() (string, error)) {
	new_scene := newExternalScene(name, NextFunc(next_func), sm.sf)
	sm.RegisterScene(new_scene)
}

// unregister Scene from scene trainsition.
// if not registered scene name is passed do nothing.
func (sm SceneManager) UnRegisterScene(name string) {
	scene, err := sm.sf.scenes.GetScene(name)
	if err != nil {
		return
	}
	sm.sf.scenes.unRegisterScene(scene)
}

// return SceneManager has the name scene.
func (sm SceneManager) SceneExists(name string) bool {
	_, err := sm.sf.scenes.GetScene(name)
	return err == nil
}
