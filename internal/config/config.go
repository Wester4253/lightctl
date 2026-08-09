package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wester4253/lightctl/go-lightctl/internal/models"
)

const (
	defaultHABaseURL = "http://100.88.207.105:8123"
	defaultEntityID  = "light.rgbic_strip_lights"
	defaultTheme     = "tokyo-night"
)

var defaultDevices = map[string]string{
	"govee":    defaultEntityID,
	"nanoleaf": "light.light_panels_55_e4_27",
}

func Path() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = "."
	}
	return filepath.Join(home, ".config", "govee", "config.json")
}

func defaultProfiles() map[string]models.Profile {
	brightness := func(value int) *int { return &value }
	return map[string]models.Profile{
		"night": {
			Name: "night", Effect: "Sleep", BrightnessPct: brightness(15), PCActionPrompt: true,
			Devices: map[string]models.Profile{"nanoleaf": {Name: "night", Effect: "moonlight", BrightnessPct: brightness(15)}},
		},
		"gaming": {
			Name: "gaming", Effect: "Rainbow", BrightnessPct: brightness(100),
			Devices: map[string]models.Profile{"nanoleaf": {Name: "gaming", Effect: "Color Burst", BrightnessPct: brightness(100)}},
		},
		"movie": {
			Name: "movie", Effect: "Candle", BrightnessPct: brightness(20),
			Devices: map[string]models.Profile{"nanoleaf": {Name: "movie", Effect: "Romantic", BrightnessPct: brightness(20)}},
		},
		"work": {
			Name: "work", BrightnessPct: brightness(100), RGBColor: []int{255, 255, 255},
			Devices: map[string]models.Profile{"nanoleaf": {Name: "work", Effect: "Triluminox Energy Crystal", BrightnessPct: brightness(100)}},
		},
		"relax": {
			Name: "relax", Effect: "Ocean", BrightnessPct: brightness(40),
			Devices: map[string]models.Profile{"nanoleaf": {Name: "relax", Effect: "Forest", BrightnessPct: brightness(40)}},
		},
	}
}

func defaultConfig() models.Config {
	return models.Config{
		HABaseURL: defaultHABaseURL,
		EntityID:  defaultEntityID,
		Theme:     defaultTheme,
		Devices:   cloneDevices(defaultDevices),
		Profiles:  defaultProfiles(),
	}
}

func cloneDevices(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for name, entity := range source {
		result[name] = entity
	}
	return result
}

func normalize(cfg models.Config) models.Config {
	if cfg.HABaseURL == "" {
		cfg.HABaseURL = defaultHABaseURL
	}
	if cfg.EntityID == "" {
		cfg.EntityID = defaultEntityID
	}
	if cfg.Theme == "" {
		cfg.Theme = defaultTheme
	}
	if cfg.Devices == nil {
		cfg.Devices = map[string]string{}
	}
	if _, ok := cfg.Devices["govee"]; !ok {
		cfg.Devices["govee"] = cfg.EntityID
	}
	// Preserve the Python implementation's compatibility behavior: a
	// configured file gets the standard nanoleaf entry unless it already
	// provides one explicitly.
	if _, ok := cfg.Devices["nanoleaf"]; !ok {
		cfg.Devices["nanoleaf"] = defaultDevices["nanoleaf"]
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]models.Profile{}
	}
	for name, profile := range cfg.Profiles {
		if profile.Name == "" {
			profile.Name = name
		}
		if profile.Devices == nil {
			profile.Devices = map[string]models.Profile{}
		}
		cfg.Profiles[name] = profile
	}
	return cfg
}

func Load() (models.Config, error) {
	path := Path()
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		cfg := defaultConfig()
		if saveErr := Save(cfg); saveErr != nil {
			return cfg, fmt.Errorf("create config: %w", saveErr)
		}
		return cfg, fmt.Errorf("created %s; add ha_token before running lightctl", path)
	}
	if err != nil {
		return models.Config{}, fmt.Errorf("could not read config at %s: %w", path, err)
	}
	var cfg models.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return models.Config{}, fmt.Errorf("could not parse config at %s: %w", path, err)
	}
	cfg = normalize(cfg)
	// Re-write in the same compatible shape so old files gain missing
	// defaults without losing custom profiles or device overrides.
	if err := Save(cfg); err != nil {
		return cfg, fmt.Errorf("could not save config at %s: %w", path, err)
	}
	if cfg.HAToken == "" {
		return cfg, fmt.Errorf("no Home Assistant token set. Add ha_token to %s", path)
	}
	return cfg, nil
}

func Save(cfg models.Config) error {
	cfg = normalize(cfg)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	path := Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func DefaultProfiles() map[string]models.Profile {
	return defaultProfiles()
}
