package csv

import (
	"testing"
)

var validGameBasePath = "../../stub/CSV/_GameBase.csv"

func TestNewGameBase(t *testing.T) {
	base, err := newGameBase(validGameBasePath)
	if err != nil {
		t.Errorf("failed to load GameBase from %v: err %v", validGameBasePath, err)
	}

	if expect := "1234-5678"; base.Code != expect {
		t.Errorf("differenct Code, expect: %v, got: %v", expect, base.Code)
	}
	if expect := int32(10); base.Version != expect {
		t.Errorf("differenct Version, expect: %v, got: %v", expect, base.Version)
	}
	if expect := "erago"; base.Title != expect {
		t.Errorf("differenct Title, expect: %v, got: %v", expect, base.Title)
	}
	if expect := "developer"; base.Author != expect {
		t.Errorf("differenct Author, expect: %v, got: %v", expect, base.Author)
	}
	if expect := "1192"; base.Develop != expect {
		t.Errorf("differenct Develop, expect: %v, got: %v", expect, base.Develop)
	}
	if expect := true; base.AllowDiffVer != expect {
		t.Errorf("differenct AllowDiffVer, expect: %v, got: %v", expect, base.AllowDiffVer)
	}
}

var validNumberConstantsPath = "../../stub/CSV/" + numberConstantsFileName

func TestNewNumberConstants(t *testing.T) {
	_, err := newNumberConstants(validNumberConstantsPath)
	if err != nil {
		t.Errorf("failed to load NumberConstants from %v: err %v", validNumberConstantsPath, err)
	}
}
