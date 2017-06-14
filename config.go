package erago

import (
	"path/filepath"

	"local/erago/flow/scene"
	"local/erago/flow/script"
	"local/erago/state"
	"local/erago/state/csv"
)

// by default, use current dir of running main.
const DefaultBaseDir = "./"

// Config holds parameters associating with Game running.
// It might be constructed by NewConfig, not Config{}.
type Config struct {
	SceneConfig  scene.Config  `toml:"scene"`
	StateConfig  state.Config  `toml:"state"`
	CSVConfig    csv.Config    `toml:"csv"`
	ScriptConfig script.Config `toml:"script"`
}

const (
	DefaultSaveFileDir     = "sav"
	DefaultCSVDir          = "CSV"
	DefaultCSVCharaPattern = "Chara/Chara*"
	DefaultScriptDir       = "ELA"
)

// construct default Config with the base directory.
func NewConfig(baseDir string) Config {
	return Config{
		SceneConfig: scene.Config{
			CanAutoSave: true,
		},
		StateConfig: state.Config{
			SaveFileDir: filepath.Join(baseDir, DefaultSaveFileDir),
		},
		ScriptConfig: script.Config{
			LoadDir:             filepath.Join(baseDir, DefaultScriptDir),
			LoadPattern:         script.LoadPattern,
			CallStackSize:       script.CallStackSize,
			RegistrySize:        script.RegistrySize,
			IncludeGoStackTrace: true,
		},
		CSVConfig: csv.Config{
			Dir:          filepath.Join(baseDir, DefaultCSVDir),
			CharaPattern: DefaultCSVCharaPattern,
		},
	}
}
