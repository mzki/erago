package stub

import (
	"path"
	"path/filepath"
	"runtime"

	"local/erago/state"
	"local/erago/state/csv"
)

// __FILE__
func GetCurrentFile() string {
	_, filename, _, _ := runtime.Caller(1)
	return filename
}

// __DIR__
func GetCurrentDir() string {
	fname := GetCurrentFile()
	return path.Dir(fname)
}

var csvDB *csv.CsvManager

// return global test csv manager which contains
// some values already.
func GetCSV() (*csv.CsvManager, error) {
	if csvDB == nil {
		csvDB = &csv.CsvManager{}
		err := csvDB.Initialize(csv.Config{
			Dir:          filepath.Join(GetCurrentDir(), "CSV"),
			CharaPattern: "Chara/Chara*",
		})
		return csvDB, err
	}
	return csvDB, nil
}

var gameState *state.GameState

// return global test game state which contains
// some values already.
func GetGameState() (*state.GameState, error) {
	csvm, err := GetCSV()
	if err != nil {
		return nil, err
	}

	if gameState == nil {
		config := state.Config{
			SaveFileDir: filepath.Join(GetCurrentDir(), "sav"),
		}
		gameState = state.NewGameState(csvm, config)
	}
	return gameState, nil
}
