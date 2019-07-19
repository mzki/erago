package repo

import (
	"os"
	"testing"

	"github.com/mzki/erago/state"
	"github.com/mzki/erago/state/csv"
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

var (
	// Global state of CSV.
	CSVDB *csv.CsvManager

	Repo state.Repository
)

func initialize() error {
	CSVDB = &csv.CsvManager{}
	err := CSVDB.Initialize(csv.Config{
		Dir:          "../../stub/CSV",
		CharaPattern: "Chara/Chara*",
	})
	if err != nil {
		return err
	}

	Repo = NewFileRepository(CSVDB, Config{
		SaveFileDir: "../../stub/sav",
	})
	return nil
}

func TestFileRepositoryImplementsInterface(t *testing.T) {
	var repo state.Repository = &FileRepository{}
	_ = repo
}

func TestMarshall(t *testing.T) {
	gamestate := state.NewGameState(CSVDB, Repo)

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

func TestMarshallWithAddChara(t *testing.T) {
	gamestate := state.NewGameState(CSVDB, Repo)

	// find first characterID
	var charaID int64 = -1
	for id, _ := range CSVDB.CharaMap {
		charaID = id
		break // just get firstID
	}
	if charaID < 0 {
		t.Fatalf("csv has no character IDs, cant test this case")
	}

	// add new character and modify its parameter.
	addedChara, err := gamestate.SystemData.Chara.AddID(charaID)
	if err != nil {
		t.Fatal(err)
	}

	const VarName = "Base"
	const IndexKey = "体力"
	const SetValue = 100

	if base, ok := addedChara.UserVariables.GetInt(VarName); !ok {
		t.Fatalf("character value %v is not found", VarName)
	} else if base.SetByStr(IndexKey, SetValue); err != nil {
		t.Fatal(err)
	}

	// save current state.
	if err := gamestate.SaveSystem(0); err != nil {
		t.Fatal(err)
	}

	// load state into empty one.
	loadedGameState := state.NewGameState(CSVDB, Repo)
	if err := loadedGameState.LoadSystem(0); err != nil {
		t.Fatal(err)
	}

	// check identity
	loadedChara := loadedGameState.SystemData.Chara.Get(0)
	loadedBase, ok := loadedChara.UserVariables.GetInt(VarName)
	if !ok {
		t.Fatalf("character value %s is not found after load state into empty one", VarName)
	}
	if v, ok := loadedBase.GetByStr(IndexKey); !ok {
		t.Errorf("cant get %s[%s] after load state into empty one", VarName, IndexKey)
		t.Log(v)
	} else if v != SetValue {
		t.Errorf("violates identity for loaded value, expect %v, got %v", SetValue, v)
	}
}

func TestMarshallWithTarget(t *testing.T) {
	gamestate := state.NewGameState(CSVDB, Repo)

	// find first characterID
	var charaID int64 = -1
	for id, _ := range CSVDB.CharaMap {
		charaID = id
		break // just get firstID
	}
	if charaID < 0 {
		t.Fatalf("csv has no character IDs, cant test this case")
	}

	// add new character and modify its parameter.
	addedChara, err := gamestate.SystemData.Chara.AddID(charaID)
	if err != nil {
		t.Fatal(err)
	}

	if err := gamestate.SystemData.Target.Set(0, addedChara); err != nil {
		t.Fatal(err)
	}

	// save current state.
	if err := gamestate.SaveSystem(0); err != nil {
		t.Fatal(err)
	}

	// load state into empty one.
	loadedGameState := state.NewGameState(CSVDB, Repo)
	if err := loadedGameState.LoadSystem(0); err != nil {
		t.Fatal(err)
	}

	// check identity
	loadedChara := loadedGameState.SystemData.Chara.Get(0)
	loadedTarget := loadedGameState.SystemData.Target.GetChara(0)
	if loadedTarget == nil {
		t.Fatal("target has no character reference!")
	}

	if loadedChara != loadedTarget {
		t.Fatal("relationship of charachter and chara reference is broken.")
	}
}
