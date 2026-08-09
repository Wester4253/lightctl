package ha

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/wester4253/lightctl/go-lightctl/internal/models"
)

type Client struct {
	cfg  models.Config
	http *http.Client
}

func NewClient(cfg models.Config) *Client {
	return &Client{cfg: cfg, http: &http.Client{Timeout: 5 * time.Second}}
}

func (c *Client) request(method, path string, body any) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode Home Assistant request: %w", err)
		}
		reader = bytes.NewReader(data)
	}
	url := strings.TrimRight(c.cfg.HABaseURL, "/") + path
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, fmt.Errorf("create Home Assistant request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.HAToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach Home Assistant at %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("Home Assistant rejected the request; check your ha_token")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Home Assistant returned HTTP %d for %s", resp.StatusCode, path)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read Home Assistant response: %w", err)
	}
	return data, nil
}

func (c *Client) turnOn(entity string, profile models.Profile) error {
	body := map[string]any{"entity_id": entity}
	if profile.Effect != "" {
		body["effect"] = profile.Effect
	}
	if profile.BrightnessPct != nil {
		body["brightness_pct"] = *profile.BrightnessPct
	}
	if len(profile.RGBColor) > 0 {
		body["rgb_color"] = profile.RGBColor
	}
	_, err := c.request(http.MethodPost, "/api/services/light/turn_on", body)
	return err
}

func (c *Client) turnOff(entity string) error {
	_, err := c.request(http.MethodPost, "/api/services/light/turn_off", map[string]string{"entity_id": entity})
	return err
}

func (c *Client) devices() map[string]string {
	if len(c.cfg.Devices) > 0 {
		return c.cfg.Devices
	}
	return map[string]string{"govee": c.cfg.EntityID}
}

func (c *Client) ApplyProfile(name string) (models.Profile, error) {
	profile, ok := c.cfg.Profiles[name]
	if !ok {
		return models.Profile{}, fmt.Errorf("no profile named %q. Available: %s", name, c.ProfileNames())
	}
	for device, entity := range c.devices() {
		if err := c.turnOn(entity, profile.ForDevice(device)); err != nil {
			return profile, err
		}
	}
	return profile, nil
}

func (c *Client) ProfileNames() string {
	names := make([]string, 0, len(c.cfg.Profiles))
	for name := range c.cfg.Profiles {
		names = append(names, name)
	}
	sortStrings(names)
	if len(names) == 0 {
		return "(none configured)"
	}
	return strings.Join(names, ", ")
}

func (c *Client) SetBrightness(value int) error {
	if value < 0 || value > 100 {
		return fmt.Errorf("brightness must be 0-100, got %d", value)
	}
	return c.applyAll(models.Profile{BrightnessPct: &value})
}

func (c *Client) SetColor(red, green, blue int) error {
	for label, value := range map[string]int{"red": red, "green": green, "blue": blue} {
		if value < 0 || value > 255 {
			return fmt.Errorf("%s must be 0-255, got %d", label, value)
		}
	}
	return c.applyAll(models.Profile{RGBColor: []int{red, green, blue}})
}

func (c *Client) SetEffect(effect string) error {
	if strings.TrimSpace(effect) == "" {
		return fmt.Errorf("effect name cannot be empty")
	}
	matched := false
	for _, entity := range c.devices() {
		state, err := c.GetState(entity)
		if err != nil {
			return err
		}
		if len(state.AvailableEffects) > 0 && !contains(state.AvailableEffects, effect) {
			continue
		}
		if err := c.turnOn(entity, models.Profile{Effect: effect}); err != nil {
			return err
		}
		matched = true
	}
	if !matched {
		return fmt.Errorf("unknown effect %q on configured lights", effect)
	}
	return nil
}

func (c *Client) SetDeviceEffect(device, effect string) error {
	entity, ok := c.devices()[device]
	if !ok {
		return fmt.Errorf("unknown device %q", device)
	}
	state, err := c.GetState(entity)
	if err != nil {
		return err
	}
	if len(state.AvailableEffects) > 0 && !contains(state.AvailableEffects, effect) {
		return fmt.Errorf("unknown effect %q for %s. Available: %s", effect, device, strings.Join(state.AvailableEffects, ", "))
	}
	return c.turnOn(entity, models.Profile{Effect: effect})
}

func (c *Client) applyAll(profile models.Profile) error {
	for _, entity := range c.devices() {
		if err := c.turnOn(entity, profile); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) TurnOn() error { return c.applyAll(models.Profile{}) }
func (c *Client) TurnOff() error {
	for _, entity := range c.devices() {
		if err := c.turnOff(entity); err != nil {
			return err
		}
	}
	return nil
}

type stateResponse struct {
	State      string `json:"state"`
	Attributes struct {
		Brightness *int     `json:"brightness"`
		Effect     string   `json:"effect"`
		EffectList []string `json:"effect_list"`
	} `json:"attributes"`
}

func (c *Client) GetState(entity string) (models.LightState, error) {
	data, err := c.request(http.MethodGet, "/api/states/"+entity, nil)
	if err != nil {
		return models.LightState{}, err
	}
	var response stateResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return models.LightState{}, fmt.Errorf("invalid state returned for %s: %w", entity, err)
	}
	var brightness *int
	if response.Attributes.Brightness != nil {
		value := int(math.Round(float64(*response.Attributes.Brightness) / 255 * 100))
		brightness = &value
	}
	return models.LightState{
		IsOn:             response.State == "on",
		BrightnessPct:    brightness,
		Effect:           response.Attributes.Effect,
		AvailableEffects: response.Attributes.EffectList,
	}, nil
}

func (c *Client) States() (map[string]models.LightState, error) {
	states := make(map[string]models.LightState, len(c.devices()))
	for name, entity := range c.devices() {
		state, err := c.GetState(entity)
		if err != nil {
			return nil, fmt.Errorf("read %s (%s): %w", name, entity, err)
		}
		states[name] = state
	}
	return states, nil
}

func (c *Client) Effects() (map[string][]string, error) {
	states, err := c.States()
	if err != nil {
		return nil, err
	}
	effects := make(map[string][]string, len(states))
	for name, state := range states {
		effects[name] = state.AvailableEffects
	}
	return effects, nil
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}
