package stub

import (
	"context"
	"fmt"
)

// implements scene.Scripter
type sceneScripter struct{}

func NewSceneScripter() *sceneScripter {
	return &sceneScripter{}
}

func (ss sceneScripter) EraCall(str string) error {
	_, err := fmt.Printf("scripter calls %s()\n", str)
	return err
}

func (ss sceneScripter) EraCallBoolArgInt(str string, n int64) (bool, error) {
	_, err := fmt.Printf("scripter calls bool = %s(%d)\n", str, n)
	return false, err
}

func (ss sceneScripter) HasEraValue(str string) bool {
	return false
}

func (ss sceneScripter) SetContext(ctx context.Context) {
}
