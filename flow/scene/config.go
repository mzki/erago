package scene

// Scene Configure
type SceneConfig struct {
	CanAutoSave bool `toml: "autosave"`
}

func NewSceneConfig() SceneConfig {
	return SceneConfig{
		CanAutoSave: true,
	}
}
