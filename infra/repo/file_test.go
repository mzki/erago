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
