package app

import (
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wester4253/lightctl/go-lightctl/internal/config"
	"github.com/wester4253/lightctl/go-lightctl/internal/ha"
	"github.com/wester4253/lightctl/go-lightctl/internal/models"
)

// Styles
var (
	// Colors
	primary   = lipgloss.Color("#8BE9FD")
	secondary = lipgloss.Color("#BD93F9")
	success   = lipgloss.Color("#50FA7B")
	warning   = lipgloss.Color("#F1FA8C")
	errorCol  = lipgloss.Color("#FF5555")
	muted     = lipgloss.Color("#6272A4")
	panel     = lipgloss.Color("#44475A")

	// Base styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primary).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(muted).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primary).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(panel).
			Padding(0, 1)

	statusPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(panel).
				Padding(1, 2).
				MarginBottom(1)

	mainPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(panel).
			Padding(1, 2)

	successDot  = lipgloss.NewStyle().Foreground(success).Render("●")
	errorDot    = lipgloss.NewStyle().Foreground(errorCol).Render("●")
	mutedText   = lipgloss.NewStyle().Foreground(muted)
	successText = lipgloss.NewStyle().Foreground(success)
	errorText   = lipgloss.NewStyle().Foreground(errorCol)
	warningText = lipgloss.NewStyle().Foreground(warning)
)

type screen int

const (
	screenMain screen = iota
	screenControl
	screenEffects
	screenBrightness
	screenColor
	screenTheme
	screenSettings
	screenPCAction
)

type menuItem struct {
	key, title, desc string
}

func (i menuItem) FilterValue() string { return i.title }
func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return i.desc }

type effectItem struct {
	device, effect string
}

func (i effectItem) FilterValue() string { return i.device + " " + i.effect }
func (i effectItem) Title() string       { return i.device + ": " + i.effect }
func (i effectItem) Description() string { return "" }

type themeItem string

func (i themeItem) FilterValue() string { return string(i) }
func (i themeItem) Title() string       { return string(i) }
func (i themeItem) Description() string { return "" }

// Messages
type statesMsg struct {
	states map[string]models.LightState
	err    error
}

type actionMsg struct {
	message string
	err     error
	prompt  bool
}

type effectsLoadedMsg struct {
	effects map[string][]string
	err     error
}

type tickMsg time.Time

type TUI struct {
	cfg           models.Config
	client        *ha.Client
	currentScreen screen
	width         int
	height        int

	// State
	states        map[string]models.LightState
	statusMessage string
	loading       bool

	// Components
	mainList list.Model
	subList  list.Model
	input    textinput.Model

	// For inputs
	brightnessValue int
}

func NewTUI(cfg models.Config, client *ha.Client) *TUI {
	// Build main menu items
	items := []list.Item{
		menuItem{"control", "Control Panel", "Live overview of all devices"},
		menuItem{"night", "Night Mode", "Apply night profile"},
		menuItem{"gaming", "Gaming Mode", "Apply gaming profile"},
		menuItem{"movie", "Movie Mode", "Apply movie profile"},
		menuItem{"work", "Work Mode", "Apply work profile"},
		menuItem{"relax", "Relax Mode", "Apply relax profile"},
		menuItem{"effects", "Effects", "Browse and select effects"},
		menuItem{"brightness", "Brightness", "Set brightness 0-100"},
		menuItem{"colors", "Colors", "Set RGB color"},
		menuItem{"power", "Power", "Toggle lights on/off"},
		menuItem{"theme", "Theme", "Change TUI theme"},
		menuItem{"settings", "Settings", "View configuration"},
	}

	delegate := list.NewDefaultDelegate()
	delegate.SetSpacing(1)
	mainList := list.New(items, delegate, 0, 0)
	mainList.Title = "lightctl"
	mainList.SetShowStatusBar(false)
	mainList.SetFilteringEnabled(false)
	mainList.Styles.Title = titleStyle

	subList := list.New([]list.Item{}, delegate, 0, 0)
	subList.SetShowStatusBar(false)
	subList.Styles.Title = titleStyle

	return &TUI{
		cfg:             cfg,
		client:          client,
		currentScreen:   screenMain,
		mainList:        mainList,
		subList:         subList,
		statusMessage:   "Loading...",
		loading:         true,
		brightnessValue: 50,
	}
}

func (m *TUI) Init() tea.Cmd {
	return tea.Batch(
		m.refreshStates(),
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *TUI) refreshStates() tea.Cmd {
	return func() tea.Msg {
		states, err := m.client.States()
		return statesMsg{states: states, err: err}
	}
}

func (m *TUI) performAction(action func() error, successMsg string, withRefresh bool, pcPrompt bool) tea.Cmd {
	return func() tea.Msg {
		err := action()
		return actionMsg{message: successMsg, err: err, prompt: pcPrompt}
	}
}

func (m *TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	// Update active component FIRST so it receives all messages
	var cmd tea.Cmd
	switch m.currentScreen {
	case screenMain, screenControl, screenSettings:
		m.mainList, cmd = m.mainList.Update(msg)
		cmds = append(cmds, cmd)
	case screenEffects, screenTheme:
		m.subList, cmd = m.subList.Update(msg)
		cmds = append(cmds, cmd)
	case screenBrightness, screenColor:
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}
	
	// Then handle our own logic
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.mainList.SetSize(msg.Width-4, msg.Height-12)
		m.subList.SetSize(msg.Width-4, msg.Height-12)
		
	case tickMsg:
		cmds = append(cmds, tickCmd())
		
	case statesMsg:
		m.loading = false
		if msg.err != nil {
			m.statusMessage = "Error: " + msg.err.Error()
		} else {
			m.states = msg.states
			m.statusMessage = "Ready"
		}
		
	case actionMsg:
		if msg.err != nil {
			m.statusMessage = "Error: " + msg.err.Error()
		} else {
			m.statusMessage = msg.message
			if msg.prompt {
				m.currentScreen = screenPCAction
			} else {
				m.currentScreen = screenMain
				cmds = append(cmds, m.refreshStates())
			}
		}
		
	case effectsLoadedMsg:
		if msg.err != nil {
			m.statusMessage = "Error loading effects: " + msg.err.Error()
			m.currentScreen = screenMain
		} else {
			items := []list.Item{}
			devices := []string{}
			for device := range msg.effects {
				devices = append(devices, device)
			}
			sort.Strings(devices)
			
			for _, device := range devices {
				effects := msg.effects[device]
				sort.Strings(effects)
				for _, effect := range effects {
					items = append(items, effectItem{device: device, effect: effect})
				}
			}
			
			m.subList = list.New(items, list.NewDefaultDelegate(), m.width-4, m.height-12)
			m.subList.Title = "Effects"
			m.subList.SetShowStatusBar(false)
			m.subList.Styles.Title = titleStyle
			m.currentScreen = screenEffects
			m.statusMessage = "Select an effect"
		}
		
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		
		if msg.String() == "q" && m.currentScreen == screenMain {
			return m, tea.Quit
		}
		
		// Handle key-specific actions
		if keyCmd := m.handleKeyPress(msg); keyCmd != nil {
			cmds = append(cmds, keyCmd)
		}
	}
	
	return m, tea.Batch(cmds...)
}

func (m *TUI) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch m.currentScreen {
	case screenMain:
		switch key {
		case "r":
			m.loading = true
			m.statusMessage = "Refreshing..."
			return m.refreshStates()
		case "enter", " ":
			if selected, ok := m.mainList.SelectedItem().(menuItem); ok {
				return m.activateMenuItem(selected)
			}
		}

	case screenControl:
		switch key {
		case "r":
			m.loading = true
			return m.refreshStates()
		case "esc", "q":
			m.currentScreen = screenMain
			return m.refreshStates()
		}

	case screenEffects:
		switch key {
		case "esc", "q":
			m.currentScreen = screenMain
		case "enter", " ":
			if selected, ok := m.subList.SelectedItem().(effectItem); ok {
				m.currentScreen = screenMain
				return m.performAction(
					func() error { return m.client.SetDeviceEffect(selected.device, selected.effect) },
					fmt.Sprintf("Set %s on %s", selected.effect, selected.device),
					true, false,
				)
			}
		}

	case screenBrightness:
		switch key {
		case "esc", "q":
			m.currentScreen = screenMain
		case "left":
			if m.brightnessValue > 0 {
				m.brightnessValue--
				m.input.SetValue(strconv.Itoa(m.brightnessValue))
			}
		case "right":
			if m.brightnessValue < 100 {
				m.brightnessValue++
				m.input.SetValue(strconv.Itoa(m.brightnessValue))
			}
		case "enter":
			val := strings.TrimSpace(m.input.Value())
			if v, err := strconv.Atoi(val); err == nil && v >= 0 && v <= 100 {
				m.brightnessValue = v
				m.currentScreen = screenMain
				return m.performAction(
					func() error { return m.client.SetBrightness(v) },
					fmt.Sprintf("Brightness set to %d%%", v),
					true, false,
				)
			} else {
				m.statusMessage = "Brightness must be 0-100"
			}
		}

	case screenColor:
		switch key {
		case "esc", "q":
			m.currentScreen = screenMain
		case "enter":
			val := strings.TrimSpace(m.input.Value())
			parts := strings.Fields(val)
			if len(parts) == 3 {
				r, err1 := strconv.Atoi(parts[0])
				g, err2 := strconv.Atoi(parts[1])
				b, err3 := strconv.Atoi(parts[2])
				if err1 == nil && err2 == nil && err3 == nil &&
					r >= 0 && r <= 255 && g >= 0 && g <= 255 && b >= 0 && b <= 255 {
					m.currentScreen = screenMain
					return m.performAction(
						func() error { return m.client.SetColor(r, g, b) },
						fmt.Sprintf("Color set to (%d, %d, %d)", r, g, b),
						true, false,
					)
				}
			}
			m.statusMessage = "Enter RGB values (0-255 each, space-separated)"
		}

	case screenTheme:
		switch key {
		case "esc", "q":
			m.currentScreen = screenMain
		case "enter", " ":
			if selected, ok := m.subList.SelectedItem().(themeItem); ok {
				m.cfg.Theme = string(selected)
				if err := config.Save(m.cfg); err != nil {
					m.statusMessage = "Error saving theme: " + err.Error()
				} else {
					m.statusMessage = "Theme saved: " + string(selected)
				}
				m.currentScreen = screenMain
			}
		}

	case screenSettings:
		if key == "esc" || key == "q" || key == "enter" {
			m.currentScreen = screenMain
		}

	case screenPCAction:
		switch key {
		case "1":
			m.currentScreen = screenMain
			return func() tea.Msg {
				return actionMsg{message: runPCAction("1")}
			}
		case "2":
			m.currentScreen = screenMain
			return func() tea.Msg {
				return actionMsg{message: runPCAction("2")}
			}
		case "3", "esc":
			m.currentScreen = screenMain
			m.statusMessage = "PC action ignored"
			return m.refreshStates()
		}
	}

	return nil
}

func (m *TUI) activateMenuItem(item menuItem) tea.Cmd {
	switch item.key {
	case "control":
		m.currentScreen = screenControl
		return m.refreshStates()

	case "night", "gaming", "movie", "work", "relax":
		profile, ok := m.cfg.Profiles[item.key]
		if !ok {
			m.statusMessage = "Profile not found: " + item.key
			return nil
		}
		return m.performAction(
			func() error { _, err := m.client.ApplyProfile(item.key); return err },
			"Activated "+item.key,
			true,
			profile.PCActionPrompt,
		)

	case "effects":
		m.loading = true
		m.statusMessage = "Loading effects..."
		return func() tea.Msg {
			effects, err := m.client.Effects()
			return effectsLoadedMsg{effects: effects, err: err}
		}

	case "brightness":
		m.currentScreen = screenBrightness
		m.input = newInput("Brightness (0-100)", strconv.Itoa(m.brightnessValue))
		return m.input.Focus()

	case "colors":
		m.currentScreen = screenColor
		m.input = newInput("RGB (e.g. 255 0 255)", "255 255 255")
		return m.input.Focus()

	case "power":
		// Toggle power based on first device state
		if len(m.states) > 0 {
			anyOn := false
			for _, state := range m.states {
				if state.IsOn {
					anyOn = true
					break
				}
			}
			if anyOn {
				return m.performAction(
					func() error { return m.client.TurnOff() },
					"Lights turned off",
					true, false,
				)
			} else {
				return m.performAction(
					func() error { return m.client.TurnOn() },
					"Lights turned on",
					true, false,
				)
			}
		}

	case "theme":
		items := []list.Item{}
		themes := []string{"tokyo-night", "dracula", "nord", "catppuccin", "gruvbox", "solarized"}
		for _, name := range themes {
			items = append(items, themeItem(name))
		}
		m.subList = list.New(items, list.NewDefaultDelegate(), m.width-4, m.height-12)
		m.subList.Title = "Themes"
		m.subList.SetShowStatusBar(false)
		m.subList.Styles.Title = titleStyle
		m.currentScreen = screenTheme

	case "settings":
		m.currentScreen = screenSettings
	}

	return nil
}

func newInput(prompt, value string) textinput.Model {
	input := textinput.New()
	input.Placeholder = prompt
	input.SetValue(value)
	input.CharLimit = 50
	input.Width = 40
	return input
}

func runPCAction(choice string) string {
	var command []string
	message := "PC action ignored"
	switch choice {
	case "1":
		message, command = "Shutting down...", []string{"systemctl", "poweroff"}
	case "2":
		message, command = "Suspending...", []string{"systemctl", "suspend"}
	default:
		return message
	}
	if err := exec.Command(command[0], command[1:]...).Run(); err != nil {
		return fmt.Sprintf("%s (error: %v)", message, err)
	}
	return message
}

func (m *TUI) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	switch m.currentScreen {
	case screenMain:
		return m.viewMain()
	case screenControl:
		return m.viewControl()
	case screenEffects, screenTheme:
		return m.viewList()
	case screenBrightness:
		return m.viewBrightness()
	case screenColor:
		return m.viewColor()
	case screenSettings:
		return m.viewSettings()
	case screenPCAction:
		return m.viewPCAction()
	}

	return ""
}

func (m *TUI) viewMain() string {
	// Header
	header := headerStyle.Render(" lightctl ") + " " +
		subtitleStyle.Render("Home Assistant lighting control")

	// Status panel
	statusPanel := m.renderStatusPanel()

	// Main list
	listView := m.mainList.View()

	// Footer
	footer := mutedText.Render(m.statusMessage + "  •  ↑/↓ navigate • Enter select • r refresh • q quit")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		statusPanel,
		listView,
		"",
		footer,
	)

	return mainPanelStyle.Width(m.width - 2).Render(content)
}

func (m *TUI) viewControl() string {
	header := headerStyle.Render(" Control Panel ")

	if m.loading {
		content := lipgloss.JoinVertical(lipgloss.Left, header, "", mutedText.Render("Loading device statistics..."))
		return mainPanelStyle.Width(m.width - 2).Render(content)
	}

	if len(m.states) == 0 {
		content := lipgloss.JoinVertical(lipgloss.Left, header, "", mutedText.Render("No devices configured"))
		return mainPanelStyle.Width(m.width - 2).Render(content)
	}

	total := len(m.states)
	onCount := 0
	for _, state := range m.states {
		if state.IsOn {
			onCount++
		}
	}

	summary := titleStyle.Render(fmt.Sprintf("Lighting Control Panel  •  %d/%d lights on", onCount, total))

	devices := []string{}
	for name := range m.states {
		devices = append(devices, name)
	}
	sort.Strings(devices)

	var deviceLines []string
	for _, name := range devices {
		state := m.states[name]
		entity := m.cfg.Devices[name]

		status := "OFF"
		if state.IsOn {
			status = "ON"
		}

		brightness := "unknown"
		if state.BrightnessPct != nil {
			brightness = fmt.Sprintf("%d%%", *state.BrightnessPct)
		}

		effect := "none"
		if state.Effect != "" {
			effect = state.Effect
		}

		deviceBlock := fmt.Sprintf(
			"%s  %s\n  Entity: %s\n  Effect: %s\n  Brightness: %s\n  Available effects: %d",
			lipgloss.NewStyle().Bold(true).Render(strings.Title(name)),
			status,
			entity,
			effect,
			brightness,
			len(state.AvailableEffects),
		)
		deviceLines = append(deviceLines, deviceBlock)
	}

	footer := mutedText.Render("Press r to refresh  •  Esc/q to go back")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		summary,
		"",
		strings.Join(deviceLines, "\n\n"),
		"",
		footer,
	)

	return mainPanelStyle.Width(m.width - 2).Render(content)
}

func (m *TUI) viewList() string {
	header := headerStyle.Render(" " + m.subList.Title + " ")
	footer := mutedText.Render("↑/↓ navigate • Enter select • Esc back")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		m.subList.View(),
		"",
		footer,
	)

	return mainPanelStyle.Width(m.width - 2).Render(content)
}

func (m *TUI) viewBrightness() string {
	header := headerStyle.Render(" Brightness ")

	// Progress bar
	bar := renderProgressBar(m.brightnessValue, 100, 40)

	footer := mutedText.Render("←/→ adjust • Enter apply • Esc cancel")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		bar,
		"",
		m.input.View(),
		"",
		footer,
	)

	return mainPanelStyle.Width(m.width - 2).Render(content)
}

func (m *TUI) viewColor() string {
	header := headerStyle.Render(" RGB Color ")
	footer := mutedText.Render("Enter space-separated RGB values • Esc cancel")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		m.input.View(),
		"",
		mutedText.Render("Example: 255 0 255"),
		"",
		footer,
	)

	return mainPanelStyle.Width(m.width - 2).Render(content)
}

func (m *TUI) viewSettings() string {
	header := headerStyle.Render(" Settings ")

	devices := []string{}
	for name, entity := range m.cfg.Devices {
		devices = append(devices, fmt.Sprintf("%s=%s", name, entity))
	}
	sort.Strings(devices)

	profiles := []string{}
	for name := range m.cfg.Profiles {
		profiles = append(profiles, name)
	}
	sort.Strings(profiles)

	lines := []string{
		"Config file: " + config.Path(),
		"Home Assistant: " + m.cfg.HABaseURL,
		"Entity: " + m.cfg.EntityID,
		"Devices: " + strings.Join(devices, ", "),
		"Theme: " + m.cfg.Theme,
		"Profiles: " + strings.Join(profiles, ", "),
		"",
		mutedText.Render("Edit config.json directly to change these or add profiles."),
	}

	footer := mutedText.Render("Enter/Esc to go back")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		strings.Join(lines, "\n"),
		"",
		footer,
	)

	return mainPanelStyle.Width(m.width - 2).Render(content)
}

func (m *TUI) viewPCAction() string {
	header := headerStyle.Render(" Profile Complete ")

	title := titleStyle.Render("Goodnight!")

	options := []string{
		"1) Shut down",
		"2) Suspend",
		"3) Ignore",
	}

	footer := mutedText.Render("Choose 1, 2, or 3 • Esc to ignore")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		title,
		"",
		strings.Join(options, "\n"),
		"",
		footer,
	)

	return mainPanelStyle.Width(m.width - 2).Render(content)
}

func (m *TUI) renderStatusPanel() string {
	if m.loading {
		return statusPanelStyle.Render(mutedText.Render("● loading..."))
	}

	if len(m.states) == 0 {
		return statusPanelStyle.Render(errorText.Render("● No device state available"))
	}

	devices := []string{}
	for name := range m.states {
		devices = append(devices, name)
	}
	sort.Strings(devices)

	var lines []string
	for _, name := range devices {
		state := m.states[name]

		dot := errorDot
		status := "Off"
		if state.IsOn {
			dot = successDot
			status = "Online"
		}

		brightness := "-"
		if state.BrightnessPct != nil {
			brightness = fmt.Sprintf("%d%%", *state.BrightnessPct)
		}

		effect := state.Effect
		if effect == "" {
			effect = "-"
		}

		line := fmt.Sprintf(
			"%s %s: %s   %s %s   %s %s",
			dot,
			lipgloss.NewStyle().Bold(true).Render(name),
			status,
			mutedText.Render("Effect"),
			effect,
			mutedText.Render("Brightness"),
			brightness,
		)
		lines = append(lines, line)
	}

	return statusPanelStyle.Render(strings.Join(lines, "\n"))
}

func renderProgressBar(value, maxValue, width int) string {
	if maxValue <= 0 {
		return ""
	}
	filled := value * width / maxValue
	if filled > width {
		filled = width
	}

	bar := successText.Render(strings.Repeat("█", filled)) +
		mutedText.Render(strings.Repeat("░", width-filled))

	return fmt.Sprintf("%s  %d%%", bar, value)
}
