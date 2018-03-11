package scene

import (
	"errors"
)

// Flow is interrupted so that quit immediatly.
var ErrorQuit = errors.New("quit")

// Flow is interrupted.
var ErrorInterrupt = errors.New("interrupted")

// Scene flow is interrupted so that start next scene immediatly.
var ErrorSceneNext = errors.New("go to next scene")

// TODO: interrupt signal to restart title scene.
// to implement this, all of game controller funtions returns error.
// because running script is interrupted by error defined above.
// var ErrorRestartTitle
