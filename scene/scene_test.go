package scene

import (
	"context"
	"testing"

	"local/erago/stub"
)

func buildSceneManager() *SceneManager {
	controller := stub.NewFlowGameController()
	scripter := stub.NewSceneScripter()
	config := Config{CanAutoSave: true}
	state, err := stub.GetGameState()
	if err != nil {
		panic(err)
	}
	m := NewSceneManager(controller, scripter, state, config)
	m.RegisterScene(SceneNameTitle, func() (string, error) {
		controller.PrintL("this is test printL")
		return "unkown scene name", nil
	})
	return m
}

func TestSceneManager(t *testing.T) {
	manager := buildSceneManager()
	defer manager.Free()

	ctx := context.Background()
	if err := manager.Run(ctx); err == nil {
		t.Error("must be error( not found next scene )")
	} else {
		t.Log("SceneManager.Run() returns:")
		t.Log(err)
	}

	manager.UnRegisterScene(SceneNameTitle)
	if err := manager.Run(ctx); err == nil {
		t.Error("must be error( not found next scene )")
	} else {
		t.Log("SceneManager.Run() returns:")
		t.Log(err)
	}
}

func TestSceneExists(t *testing.T) {
	m := buildSceneManager()
	defer m.Free()

	if has := m.SceneExists(SceneNameTitle); !has {
		t.Errorf("SceneManager must have the scene %s, but does not", SceneNameTitle)
	}
}
