package scene

// Representing each scene Flow.
type Scene interface {
	// Scene updates self to next Scene
	Next() (Scene, error)
	// returns self name
	Name() string
}

// common of scene
type sceneCommon struct {
	*sceneFields

	atStartScene string
	atStartEvent string
}

func newSceneCommon(name string, sf *sceneFields) sceneCommon {
	return sceneCommon{
		sceneFields:  sf,
		atStartScene: name + ScrSep + ScrScenePrefix,
		atStartEvent: name + ScrSep + ScrEventPrefix + ScrSep + "start",
	}
}

// it is called at start of Next() in every scene.
func (common sceneCommon) atStart() (Scene, error) {
	called, err := common.Script().checkCall(common.atStartScene)
	if called {
		return common.Scenes().Next(), err
	}

	err = common.Script().maybeCall(common.atStartEvent)
	return nil, err
}

func (common sceneCommon) Name() string { return "no-name" }

func (common sceneCommon) Next() (Scene, error) {
	if next, err := common.atStart(); next != nil || err != nil {
		return next, err
	}
	return common.Scenes().Next(), nil
}

// It is used to get current scene's next in scene trainsition.
// Returned string must be name of next scene, and
// error controlls scene trainsition which is defined in
// "erago/scene" package.
type NextFunc func() (string, error)

// it is used for defining user custom scene.
type externalScene struct {
	sceneName string
	nextFunc  NextFunc
	*sceneFields
}

func newExternalScene(name string, next_func NextFunc, sf *sceneFields) Scene {
	return &externalScene{
		sceneName:   name,
		nextFunc:    next_func,
		sceneFields: sf,
	}
}

func (ex externalScene) Name() string {
	return ex.sceneName
}

func (ex externalScene) Next() (Scene, error) {
	next_name, err := ex.nextFunc()
	if err != nil {
		return nil, err
	}
	return ex.scenes.GetScene(next_name)
}
