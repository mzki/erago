package erago

import (
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

// construct default Config
func NewConfig(baseDir string) Config {
	return Config{
		BaseDir:      baseDir,
		SceneConfig:  scene.NewSceneConfig(),
		StateConfig:  state.NewConfig(baseDir),
		ScriptConfig: script.NewConfig(baseDir),
		CSVConfig:    csv.NewConfig(baseDir),
	}
}

// set base directory to config. changing base directory propagates all of its fields,
// StateConfig, CSVConfig and ScriptConfig.
func (conf *Config) SetBaseDir(baseDir string) {
	conf.BaseDir = baseDir
	conf.StateConfig.SetBaseDir(baseDir)
	conf.CSVConfig.SetBaseDir(baseDir)
	conf.ScriptConfig.SetBaseDir(baseDir)
}
