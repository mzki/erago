package state

import (
	"os"
	"testing"

	"local/erago/state/csv"
)

// initialize and finalize before and after testing.
func TestMain(m *testing.M) {
	if err := initialize(); err != nil {
		panic(err)
	}

	code := m.Run()
	// finalize()
	os.Exit(code)
}

// Global state of CSV.
var CSVDB *csv.CsvManager

func initialize() error {
	CSVDB = &csv.CsvManager{}
	return CSVDB.Initialize(csv.Config{
		Dir:          "./../stub/CSV",
		CharaPattern: "Chara/Chara*",
	})
}

var stateConfig = Config{
	SaveFileDir: "../stub/sav",
}

func TestNewGameState(t *testing.T) {
	gamestate := NewGameState(CSVDB, stateConfig)

	base_vars, ok := gamestate.SystemData.GetInt("Number")
	if !ok {
		t.Error("Can not get System's Number variable.")
	}

	if _, ok := base_vars.GetByStr("数値１"); !ok {
		t.Error("Can not get Number[”数値１”] variable.")
	}

	// compare length of
	// if base_vars.Len() != CSVDB.CharaMap[1].
}

func TestMarshall(t *testing.T) {
	gamestate := NewGameState(CSVDB, stateConfig)

	number, _ := gamestate.SystemData.GetInt("Number")
	number.Set(0, 100)
	if err := gamestate.SaveSystem(0); err != nil {
		t.Fatal(err)
	}

	number.Set(0, -99)
	system_pointer_before_load := gamestate.SystemData

	if err := gamestate.LoadSystem(0); err != nil {
		t.Fatal(err)
	}

	if number.Get(0) == -99 {
		t.Error("load savefile but not reflecting values")
	}
	if system_pointer_before_load != gamestate.SystemData {
		t.Error("object pointer is changed before and after unmarshall.")
	}
}

func TestNewCharacter(t *testing.T) {
	gamestate := NewGameState(CSVDB, stateConfig)
	const CHARA_ID = 1 // it must exist

	chara, err := gamestate.SystemData.Chara.AddID(CHARA_ID)
	if err != nil {
		t.Fatal(err)
	}
	base, ok := chara.GetInt("Base")
	if !ok {
		t.Fatal("the user variable Base is not found")
	}

	base.SetByStr("体力", 100)
	hp, ok := base.GetByStr("体力")
	if !ok {
		t.Fatal("Base［体力］is not found")
	}
	if hp != 100 {
		t.Error("different setted parametrer and got parametrer in Base[\"体力\"]")
	}

	csvc, ok := gamestate.CSV.CharaMap[CHARA_ID]
	if !ok {
		t.Fatal("no exists csv Character ID: 0.")
	}
	if csvbase := csvc.GetIntMap()["Base"]; csvbase[0] == hp {
		t.Error("setting parametrer to chara influeces csv parametrer: it breaks CSV constants")
	}
}
