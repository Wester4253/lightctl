package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wester4253/lightctl/go-lightctl/internal/app"
	"github.com/wester4253/lightctl/go-lightctl/internal/config"
	"github.com/wester4253/lightctl/go-lightctl/internal/ha"
	"github.com/wester4253/lightctl/go-lightctl/internal/models"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	client := ha.NewClient(cfg)
	if len(os.Args) == 1 {
		if _, err := tea.NewProgram(app.NewTUI(cfg, client), tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if err := runCLI(os.Args[1:], cfg, client); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func runCLI(args []string, cfg models.Config, client *ha.Client) error {
	if len(args) == 0 {
		return fmt.Errorf("missing command")
	}
	switch args[0] {
	case "night", "gaming", "movie", "work", "relax":
		if len(args) != 1 {
			return fmt.Errorf("usage: %s", args[0])
		}
		return applyProfile(client, args[0])
	case "profiles":
		if len(args) != 1 {
			return fmt.Errorf("usage: profiles")
		}
		printProfileNames(cfg)
		return nil
	case "profile":
		if len(args) != 2 {
			return fmt.Errorf("usage: profile NAME (or profile names)")
		}
		if args[1] == "names" {
			printProfileNames(cfg)
			return nil
		}
		return applyProfile(client, args[1])
	case "brightness":
		if len(args) != 2 {
			return fmt.Errorf("usage: brightness 0-100")
		}
		value, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("brightness must be an integer from 0 to 100")
		}
		if err := client.SetBrightness(value); err != nil {
			return err
		}
		fmt.Printf("Brightness set to %d%%.\n", value)
		return nil
	case "color":
		if len(args) != 4 {
			return fmt.Errorf("usage: color R G B")
		}
		values := make([]int, 3)
		for i := range values {
			value, err := strconv.Atoi(args[i+1])
			if err != nil {
				return fmt.Errorf("color channels must be integers from 0 to 255")
			}
			values[i] = value
		}
		if err := client.SetColor(values[0], values[1], values[2]); err != nil {
			return err
		}
		fmt.Printf("Color set to (%d, %d, %d).\n", values[0], values[1], values[2])
		return nil
	case "effect":
		if len(args) != 2 {
			return fmt.Errorf("usage: effect NAME")
		}
		if err := client.SetEffect(args[1]); err != nil {
			return err
		}
		fmt.Printf("Effect set to '%s'.\n", args[1])
		return nil
	case "effects":
		if len(args) != 1 {
			return fmt.Errorf("usage: effects")
		}
		effects, err := client.Effects()
		if err != nil {
			return err
		}
		names := sortedKeys(effects)
		any := false
		for _, device := range names {
			if len(effects[device]) == 0 {
				continue
			}
			any = true
			fmt.Printf("%s:\n", device)
			for _, effect := range effects[device] {
				fmt.Printf("  %s\n", effect)
			}
		}
		if !any {
			fmt.Println("No effects reported.")
		}
		return nil
	case "status":
		if len(args) != 1 {
			return fmt.Errorf("usage: status")
		}
		return printStatus(client)
	case "on":
		if len(args) != 1 {
			return fmt.Errorf("usage: on")
		}
		if err := client.TurnOn(); err != nil {
			return err
		}
		fmt.Println("Light on.")
		return nil
	case "off":
		if len(args) != 1 {
			return fmt.Errorf("usage: off")
		}
		if err := client.TurnOff(); err != nil {
			return err
		}
		fmt.Println("Light off.")
		return nil
	default:
		return fmt.Errorf("unknown command %q. Try status, profiles, profile, effects, effect, brightness, color, on, or off", args[0])
	}
}

func applyProfile(client *ha.Client, name string) error {
	profile, err := client.ApplyProfile(name)
	if err != nil {
		return err
	}
	fmt.Printf("Activated '%s'.\n", name)
	if profile.PCActionPrompt {
		promptPCAction()
	}
	return nil
}

func printProfileNames(cfg models.Config) {
	names := make([]string, 0, len(cfg.Profiles))
	for name := range cfg.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) == 0 {
		fmt.Println("(no profiles configured)")
		return
	}
	for _, name := range names {
		fmt.Println(name)
	}
}

func printStatus(client *ha.Client) error {
	states, err := client.States()
	if err != nil {
		return err
	}
	names := make([]string, 0, len(states))
	for name := range states {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		state := states[name]
		status := "off"
		if state.IsOn {
			status = "on"
		}
		brightness := "-"
		if state.BrightnessPct != nil {
			brightness = fmt.Sprintf("%d%%", *state.BrightnessPct)
		}
		effect := state.Effect
		if effect == "" {
			effect = "-"
		}
		fmt.Printf("%s: %s\n  Effect: %s\n  Brightness: %s\n", name, status, effect, brightness)
	}
	return nil
}

func promptPCAction() {
	fmt.Println("\nGoodnight!")
	fmt.Println("===========")
	fmt.Println("1) Shut down")
	fmt.Println("2) Sleep")
	fmt.Println("3) Ignore")
	fmt.Print("Choose: ")
	choice, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && len(choice) == 0 {
		fmt.Println("\nIgnoring PC action.")
		return
	}
	choice = strings.TrimSpace(choice)
	message, command, valid := pcAction(choice)
	if !valid {
		fmt.Printf("Error: invalid option %q; choose 1, 2, or 3.\n", choice)
		return
	}
	if command == nil {
		fmt.Println(message)
		return
	}
	if err := exec.Command(command[0], command[1:]...).Run(); err != nil {
		fmt.Printf("Error: %s: %v\n", message, err)
		return
	}
	fmt.Println(message)
}

func pcAction(choice string) (string, []string, bool) {
	switch choice {
	case "1":
		return "Shutting down...", []string{"systemctl", "poweroff"}, true
	case "2":
		return "Sleeping...", []string{"systemctl", "suspend"}, true
	case "3":
		return "Ignoring PC action.", nil, true
	default:
		return "", nil, false
	}
}

func sortedKeys(values map[string][]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
