package script

import (
	"local/erago/scene"
)

// game controller controlls game flow, input, output, ... etc.
type GameController interface {
	scene.GameController

	// scene controll
	//
	// do multiple trains defined by commands.
	DoTrainsScene(commands []int64) error

	// starting scene of loading game data.
	DoLoadGameScene() error
	// starting scene of saving game data.
	DoSaveGameScene() error

	// set next scene specified by its name.
	SetNextSceneByName(name string) error

	// register new scene flow using its name and the function desclibeing its flow.
	// next_func must return next scene name to move to next flow.
	RegisterSceneFunc(name string, next_func func() (string, error))

	// remove registered scene specified name.
	UnRegisterScene(name string)
}
