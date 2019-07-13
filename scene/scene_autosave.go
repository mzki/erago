package scene

// AUTOSAVE SCENE
type autosaveScene struct {
	sceneCommon
}

func newAutosaveScene(sf *sceneFields) *autosaveScene {
	return &autosaveScene{newSceneCommon(SceneNameAutosave, sf)}
}

func (autosaveScene) Name() string { return SceneNameAutosave }

// +scene: autosave
// 自動保存シーンです。
// 現在のゲームの状態を自動保存します。
const (
	// +callback: {{.Name}}()
	// loadendシーン以外のシーンからbaseシーンへ遷移したとき、
	// 現在のゲームの状態を自動で保存します。
	// この関数によって、その保存処理を置き換えることができます。
	// もし、自動保存処理を行いたくない場合、この関数を定義し、
	// その中で何も処理を行わないことで実現できます。
	ScrAutoSaveReplace = "autosave_replace"
)

func (sc autosaveScene) Next() (Scene, error) {
	if next, err := sc.atStart(); next != nil || err != nil {
		return next, err
	}

	// from not load game scene, auto saving if allowed.
	prev := sc.Scenes().Prev()
	prev_is_not_load_end := prev != nil && prev.Name() != SceneNameLoadEnd

	if prev_is_not_load_end && sc.Config().CanAutoSave {
		called, err := sc.Script().checkCall(ScrAutoSaveReplace)
		if err != nil {
			return nil, err
		}

		if !called {
			// do builtin-autosave flow
			if err := saveGameSceneProcess(autoSaveNumber, sc.sceneFields); err != nil {
				return nil, err
			}
		}
	}

	if scenes := sc.Scenes(); scenes.HasNext() {
		return scenes.Next(), nil
	} else {
		return scenes.GetScene(SceneNameBase)
	}
}
