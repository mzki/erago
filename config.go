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
// It should be constructed by NewConfig, not Config{}, because
// unexported fields exist.
type Config struct {
	BaseDir string `toml:"base_dir"` // Base directory for game.

	SceneConfig  scene.SceneConfig `toml:"scene"`
	StateConfig  state.Config      `toml:"state"`
	CSVConfig    csv.Config        `toml:"csv"`
	ScriptConfig script.Config     `toml:"script"`
}

const (
	DefaultSaveFileDir     = "sav"
	DefaultCSVDir          = "CSV"
	DefaultCSVCharaPattern = "Chara/Chara*"
)

// construct default Config
func NewConfig(baseDir string) Config {
	return Config{
		BaseDir:     baseDir,
		SceneConfig: scene.NewSceneConfig(),
		StateConfig: state.Config{
			SaveFileDir: filepath.Join(baseDir, DefaultSaveFileDir),
		},
		ScriptConfig: script.NewConfig(baseDir),
		CSVConfig: csv.Config{
			Dir:          filepath.Join(baseDir, DefaultCSVDir),
			CharaPattern: DefaultCSVCharaPattern,
		},
	}
}

// set base directory to config. changing base directory propagates all of its fields,
// StateConfig, CSVConfig and ScriptConfig.
func (conf *Config) SetBaseDir(baseDir string) {
	conf.BaseDir = baseDir
	conf.ScriptConfig.SetBaseDir(baseDir)
}
