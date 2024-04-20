package stub

import (
	"github.com/mzki/erago/state"
)

// implemnts script.GameController
type scriptGameController struct {
	*state.GameState
	*sceneIOController
}

func NewScriptGameController() *scriptGameController {
	gstate, err := GetGameState()
	if err != nil {
		panic(err)
	}
	return &scriptGameController{gstate, NewFlowGameController()}
}

// GetInputQueue gets internal scriptInputQueuer from scriptGameController.
// The internal scriptInputQueuer affects the result of calling input APIs such as
// RawInputXXX and CommandXXX.
func GetInputQueue(ui *scriptGameController) *scriptInputQueuer {
	return ui.sceneIOController.scriptInputQueuer
}

func (ui scriptGameController) DoTrainsScene(cmds []int64) error                                { return nil }
func (ui scriptGameController) DoLoadGameScene() error                                          { return nil }
func (ui scriptGameController) DoSaveGameScene() error                                          { return nil }
func (ui scriptGameController) SetNextSceneByName(name string) error                            { return nil }
func (ui scriptGameController) CurrentSceneName() string                                        { return "dummy_current_snene" }
func (ui scriptGameController) NextSceneName() string                                           { return "dummy_next_snene" }
func (ui scriptGameController) RegisterSceneFunc(name string, next_func func() (string, error)) {}
func (ui scriptGameController) UnRegisterScene(name string)                                     {}
