package stub

import (
	"local/erago/state"
)

// implemnts script.GameController
type scriptGameController struct {
	*state.GameState
	*flowGameController
}

func NewScriptGameController() *scriptGameController {
	gstate, err := GetGameState()
	if err != nil {
		panic(err)
	}
	return &scriptGameController{gstate, NewFlowGameController()}
}

func (ui scriptGameController) DoTrainsScene(cmds []int64) error                                { return nil }
func (ui scriptGameController) DoLoadGameScene() error                                          { return nil }
func (ui scriptGameController) DoSaveGameScene() error                                          { return nil }
func (ui scriptGameController) SetNextSceneByName(name string) error                            { return nil }
func (ui scriptGameController) RegisterSceneFunc(name string, next_func func() (string, error)) {}
func (ui scriptGameController) UnRegisterScene(name string)                                     {}
