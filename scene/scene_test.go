package scene

import (
	"context"
	"errors"
	"testing"

	"github.com/mzki/erago/stub"
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
	m.RegisterSceneFunc(SceneNameTitle, func() (string, error) {
		controller.PrintL("this is test printL")
		return "unkown scene name", nil
	})
	return m
}

func TestSceneManager(t *testing.T) {
	manager := buildSceneManager()
	defer manager.Free()

	ctx := context.Background()
	if err := manager.Run(ctx, SceneNameTitle); err == nil {
		t.Error("must be error( not found next scene )")
	} else if errors.Is(err, ErrorRunNextSceneNotFound) {
		// intended error. ignore
	} else {
		t.Fatal(err)
	}

	manager.UnRegisterScene(SceneNameTitle)
	if err := manager.Run(ctx, SceneNameTitle); err == nil {
		t.Error("must be error( not found next scene )")
	} else {
		t.Log("SceneManager.Run() returns:")
		t.Log(err)
	}
}

func TestSceneExists(t *testing.T) {
	m := buildSceneManager()
	defer m.Free()

	// case exist
	if has := m.SceneExists(SceneNameTitle); !has {
		t.Errorf("SceneManager must have the scene %s, but does not", SceneNameTitle)
	}

	// case no exist
	m.UnRegisterScene(SceneNameTitle)
	if has := m.SceneExists(SceneNameTitle); has {
		t.Errorf("After UnRegisterScene, SceneManager must NOT have the scene %s, but does", SceneNameTitle)
	}
}

func TestSceneCurrentName(t *testing.T) {
	m := buildSceneManager()
	defer m.Free()

	for _, tt := range []struct {
		name string
		next string
	}{
		{name: "test_scene_name1", next: "test_scene_name2"},
		{name: "test_scene_name2", next: "__unknown__"},
	} {
		ttt := tt // to captrue current value into closure.
		m.RegisterSceneFunc(tt.name, func() (string, error) {
			if name := m.CurrentSceneName(); name != ttt.name {
				t.Errorf("CurrentSceneName is unmatch, expect: %v, got: %v", ttt.name, name)
			}
			return ttt.next, nil
		})
	}

	ctx := context.Background()
	err := m.Run(ctx, "test_scene_name1")
	if errors.Is(err, ErrorRunNextSceneNotFound) {
		// intended error. ignore.
	} else {
		t.Fatal(err)
	}
}

func TestSceneNextName(t *testing.T) {
	m := buildSceneManager()
	defer m.Free()

	for _, tt := range []struct {
		name              string
		next              string
		shouldNoErrAtNext bool
	}{
		{name: "test_scene_name1", next: "test_scene_name2", shouldNoErrAtNext: true},
		{name: "test_scene_name2", next: "test_scene_name3", shouldNoErrAtNext: true},
		{name: "test_scene_name3", next: "__unknown__", shouldNoErrAtNext: false},
	} {
		ttt := tt // to captrue current value into closure.
		m.RegisterSceneFunc(tt.name, func() (string, error) {
			if name := m.NextSceneName(); name != "" {
				t.Errorf("NextSceneName should empty at scene, expect: %v, got: %v", "", name)
			}
			if err := m.SetNextSceneByName(ttt.next); ttt.shouldNoErrAtNext && err != nil {
				t.Error(err)
			}
			if name := m.NextSceneName(); ttt.shouldNoErrAtNext && name != ttt.next {
				t.Errorf("NextSceneName unmatch, expect: %v, got: %v", ttt.next, name)
			} else if !ttt.shouldNoErrAtNext && name != "" {
				t.Errorf("NextSceneName should empty at next scene not found, expect:%v, got:%v", "", name)
			}
			return ttt.next, nil
		})
	}

	ctx := context.Background()
	err := m.Run(ctx, "test_scene_name1")
	if errors.Is(err, ErrorRunNextSceneNotFound) {
		// intended error. ignore.
	} else {
		t.Fatal(err)
	}
}
