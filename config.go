package erago

import (
	"path/filepath"

	"github.com/mzki/erago/infra/repo"
	"github.com/mzki/erago/infra/script"
	"github.com/mzki/erago/scene"
	"github.com/mzki/erago/state/csv"
)

// by default, use current dir of running main.
const DefaultBaseDir = "./"

// Config holds parameters associating with Game running.
// It might be constructed by NewConfig, not Config{}.
type Config struct {
	SceneConfig  scene.Config  `toml:"scene"`
	RepoConfig   repo.Config   `toml:"save"`
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
		RepoConfig: repo.Config{
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
