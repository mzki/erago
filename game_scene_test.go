package erago

import (
	"testing"

	"github.com/mzki/erago/scene"
)

func TestFillStructByMap(t *testing.T) {
	const Text = "Loaaaading"
	data := map[string]string{
		"LoadingMessage": Text,
	}

	var replaceText scene.ConfigReplaceText

	err := fillStructByMap(&replaceText, data)
	if err != nil {
		t.Fatal(err)
	}

	if got := replaceText.LoadingMessage; got != Text {
		t.Errorf("struct field is not filled, expect: %s, got: %s", Text, got)
	}
	if got := replaceText.MoneyFormat; got != "" {
		t.Errorf("uncovered struct field is filled, expect: %s, got: %s", "", got)
	}
}

func TestFillStructByMapNoEffect(t *testing.T) {
	const Text = "Loaaaading"
	data := map[string]string{
		"loading_message": Text,
	}

	var replaceText scene.ConfigReplaceText

	err := fillStructByMap(&replaceText, data)
	if err != nil {
		t.Fatal(err)
	}

	if got := replaceText.LoadingMessage; got != "" {
		t.Errorf("struct field should be empty, got: %s", got)
	}
}
