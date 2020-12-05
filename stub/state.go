package stub

import (
	"path"
	"path/filepath"
	"runtime"

	"github.com/mzki/erago/infra/repo"
	"github.com/mzki/erago/state"
	"github.com/mzki/erago/state/csv"
)

// GetCurrentFile is same as __FILE__ in c macro
func GetCurrentFile() string {
	_, filename, _, _ := runtime.Caller(1)
	return filename
}

// GetCurrentDir is same as __DIR__ in c macro
func GetCurrentDir() string {
	fname := GetCurrentFile()
	return path.Dir(fname)
}

var (
	csvDB      *csv.CsvManager
	olderCsvDB *csv.CsvManager
)

const olderSubDir = "olderData"

// GetCSV returns global test csv manager which contains
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

// GetOlderCSV returns global test csv manager with older data schema.
func GetOlderCSV() (*csv.CsvManager, error) {
	if olderCsvDB == nil {
		olderCsvDB = &csv.CsvManager{}
		err := olderCsvDB.Initialize(csv.Config{
			Dir:          filepath.Join(GetCurrentDir(), olderSubDir, "CSV"),
			CharaPattern: "Chara/Chara*",
		})
		return olderCsvDB, err
	}
	return olderCsvDB, nil
}

var (
	gameState      *state.GameState
	olderGameState *state.GameState
)

// GetGameState returns global test game state which contains
// some values already.
func GetGameState() (*state.GameState, error) {
	csvm, err := GetCSV()
	if err != nil {
		return nil, err
	}

	if gameState == nil {
		config := repo.Config{
			SaveFileDir: filepath.Join(GetCurrentDir(), "sav"),
		}
		gameState = state.NewGameState(csvm, repo.NewFileRepository(csvm, config))
	}
	return gameState, nil
}

// GetOlderGameState returns global test game state which contains
// older data schema.
func GetOlderGameState() (*state.GameState, error) {
	csvm, err := GetOlderCSV()
	if err != nil {
		return nil, err
	}

	if olderGameState == nil {
		config := repo.Config{
			// same as normal save data directory so that older and newer save data are shared on this point.
			SaveFileDir: filepath.Join(GetCurrentDir(), "sav"),
		}
		olderGameState = state.NewGameState(csvm, repo.NewFileRepository(csvm, config))
	}
	return olderGameState, nil
}
