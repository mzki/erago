package scene

// Scene Configure
type SceneConfig struct {
	CanAutoSave bool `toml:"can_autosave"`
}

func NewSceneConfig() SceneConfig {
	return SceneConfig{
		CanAutoSave: true,
	}
}
