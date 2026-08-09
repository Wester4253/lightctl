package models

type Profile struct {
	Name           string             `json:"name"`
	Effect         string             `json:"effect,omitempty"`
	BrightnessPct  *int               `json:"brightness_pct,omitempty"`
	RGBColor       []int              `json:"rgb_color,omitempty"`
	PCActionPrompt bool               `json:"pc_action_prompt,omitempty"`
	Devices        map[string]Profile `json:"devices,omitempty"`
}

type Config struct {
	HABaseURL string             `json:"ha_base_url"`
	HAToken   string             `json:"ha_token"`
	EntityID  string             `json:"entity_id"`
	Devices   map[string]string  `json:"devices"`
	Theme     string             `json:"theme"`
	Profiles  map[string]Profile `json:"profiles"`
}

type LightState struct {
	IsOn             bool
	BrightnessPct    *int
	Effect           string
	AvailableEffects []string
}

func (s LightState) EffectOrNone() string {
	if s.Effect == "" {
		return "-"
	}
	return s.Effect
}

func (p Profile) ForDevice(device string) Profile {
	if override, ok := p.Devices[device]; ok {
		return override
	}
	return p
}
