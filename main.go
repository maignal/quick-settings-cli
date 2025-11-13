package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
)

type model struct {
	menuItems             []string
	audioOutputs          []string
	bluetoothDevices      []string
	cursor                int
	audioOutputsVisible   bool
	bluetoothDevicesVisible bool
	airplaneModeEnabled   bool // Added
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc":
			if m.audioOutputsVisible {
				m.audioOutputsVisible = false
				m.cursor = 0 // "Audio Output" is at index 0
			} else if m.bluetoothDevicesVisible {
				m.bluetoothDevicesVisible = false
				m.cursor = 1 // "Bluetooth" is at index 1
			} else {
				return m, tea.Quit
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			totalItems := len(m.menuItems)
			if m.audioOutputsVisible {
				totalItems += len(m.audioOutputs)
			}
			if m.bluetoothDevicesVisible {
				totalItems += len(m.bluetoothDevices)
			}
			if m.cursor < totalItems-1 {
				m.cursor++
			}

		case "l": // Expand
			audioItemIndex := 0
			bluetoothItemIndex := 1 // Base index in menuItems
			if m.audioOutputsVisible {
				bluetoothItemIndex += len(m.audioOutputs)
			}

			if m.cursor == audioItemIndex { // On "Audio Output"
				m.audioOutputsVisible = true
				m.bluetoothDevicesVisible = false // Collapse other
				if len(m.audioOutputs) > 0 {
					m.cursor = audioItemIndex + 1
				}
			} else if m.cursor == bluetoothItemIndex { // On "Bluetooth"
				m.bluetoothDevicesVisible = true
				m.audioOutputsVisible = false // Collapse other
				if len(m.bluetoothDevices) > 0 {
					m.cursor = bluetoothItemIndex + 1
				}
			}

		case "h": // Collapse
			audioItemIndex := 0
			bluetoothItemIndex := 1 // Base index in menuItems
			if m.audioOutputsVisible {
				bluetoothItemIndex += len(m.audioOutputs)
			}

			if m.audioOutputsVisible && m.cursor > audioItemIndex && m.cursor <= audioItemIndex+len(m.audioOutputs) {
				m.audioOutputsVisible = false
				m.cursor = audioItemIndex
			} else if m.bluetoothDevicesVisible && m.cursor > bluetoothItemIndex && m.cursor <= bluetoothItemIndex+len(m.bluetoothDevices) {
				m.bluetoothDevicesVisible = false
				m.cursor = 1
			}

		case "enter", " ":
			audioItemIndex := 0
			bluetoothItemIndex := 1 // Base index in menuItems
			if m.audioOutputsVisible {
				bluetoothItemIndex += len(m.audioOutputs)
			}

			// Calculate Airplane Mode index
			airplaneModeItemIndex := 2 // Base index
			if m.audioOutputsVisible {
				airplaneModeItemIndex += len(m.audioOutputs)
			}
			if m.bluetoothDevicesVisible {
				airplaneModeItemIndex += len(m.bluetoothDevices)
			}

			if m.cursor == audioItemIndex { // "Audio Output" is selected
				m.audioOutputsVisible = !m.audioOutputsVisible
				m.bluetoothDevicesVisible = false // Collapse other
				if m.audioOutputsVisible && len(m.audioOutputs) > 0 {
					m.cursor = audioItemIndex + 1
				} else {
					m.cursor = audioItemIndex
				}
			} else if m.cursor == bluetoothItemIndex { // "Bluetooth" is selected
				m.bluetoothDevicesVisible = !m.bluetoothDevicesVisible
				m.audioOutputsVisible = false // Collapse other
				if m.bluetoothDevicesVisible && len(m.bluetoothDevices) > 0 {
					m.cursor = bluetoothItemIndex + 1
				} else {
					m.cursor = bluetoothItemIndex
				}
			} else if m.cursor == airplaneModeItemIndex { // "Airplane Mode" is selected
				m.airplaneModeEnabled = !m.airplaneModeEnabled

				// Collapse other menus
				m.audioOutputsVisible = false
				m.bluetoothDevicesVisible = false

				// Run the toggle command
				var cmd *exec.Cmd
				if m.airplaneModeEnabled {
					cmd = exec.Command("nmcli", "radio", "all", "off")
				} else {
					cmd = exec.Command("nmcli", "radio", "all", "on")
				}
				cmd.Run() // Fire and forget

			} else if m.audioOutputsVisible && m.cursor > audioItemIndex && m.cursor <= audioItemIndex+len(m.audioOutputs) {
				// An audio device is selected
				return m, tea.Quit
			} else if m.bluetoothDevicesVisible && m.cursor > bluetoothItemIndex && m.cursor <= bluetoothItemIndex+len(m.bluetoothDevices) {
				// A bluetooth device is selected
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	s := "Quick Settings\n\n"

	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	statusStyle := lipgloss.NewStyle().Faint(true)

	currentItemIndex := 0

	// Render main menu
	for _, item := range m.menuItems {
		cursor := " "
		if m.cursor == currentItemIndex {
			cursor = ">"
		}
		line := fmt.Sprintf("%s %s", cursor, item)

		// Add status for Airplane Mode
		if item == "Airplane Mode" {
			status := "[Off]"
			if m.airplaneModeEnabled {
				status = "[On]"
			}
			line = fmt.Sprintf("%s %s %s", cursor, item, statusStyle.Render(status))
		}

		if m.cursor == currentItemIndex {
			s += selectedStyle.Render(line) + "\n"
		} else {
			s += line + "\n"
		}

		currentItemIndex++ // Increment for the main menu item

		// Render audio outputs if visible
		if item == "Audio Output" && m.audioOutputsVisible {
			for _, audioDevice := range m.audioOutputs {
				audioCursor := " "
				if m.cursor == currentItemIndex {
					audioCursor = ">"
				}
				audioLine := fmt.Sprintf("  %s %s", audioCursor, audioDevice)
				if m.cursor == currentItemIndex {
					s += selectedStyle.Render(audioLine) + "\n"
				} else {
					s += audioLine + "\n"
				}
				currentItemIndex++
			}
		}

		// Render bluetooth devices if visible
		if item == "Bluetooth" && m.bluetoothDevicesVisible {
			for _, btDevice := range m.bluetoothDevices {
				btCursor := " "
				if m.cursor == currentItemIndex {
					btCursor = ">"
				}
				btLine := fmt.Sprintf("  %s %s", btCursor, btDevice)
				if m.cursor == currentItemIndex {
					s += selectedStyle.Render(btLine) + "\n"
				} else {
					s += btLine + "\n"
				}
				currentItemIndex++
			}
		}
	}

	s += "\nPress l/enter to expand, h/esc to collapse, q to quit.\n"
	return s
}

// getBluetoothDevices runs `bluetoothctl devices` and parses the output
func getBluetoothDevices() ([]string, error) {
	cmd := exec.Command("bluetoothctl", "devices")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("could not run bluetoothctl: %w. Is bluetooth service running?", err)
	}

	lines := strings.Split(string(out), "\n")
	var devices []string
	for _, line := range lines {
		if line != "" {
			parts := strings.Split(line, " ")
			if len(parts) > 2 {
				deviceName := strings.Join(parts[2:], " ")
				devices = append(devices, deviceName)
			}
		}
	}
	return devices, nil
}

// getAirplaneModeStatus checks if 'nmcli radio wifi' reports 'disabled'
func getAirplaneModeStatus() (bool, error) {
	cmd := exec.Command("nmcli", "radio", "wifi")
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("could not check nmcli: %w. Is NetworkManager installed?", err)
	}

	// Output is "enabled" or "disabled"
	if strings.Contains(string(out), "disabled") {
		return true, nil // Radios are off, so Airplane Mode is ON
	}

	return false, nil // Radios are on, so Airplane Mode is OFF
}

func main() {
	// Get Audio Devices
	cmd := exec.Command("pactl", "list", "short", "sinks")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("could not get audio devices:", err)
		os.Exit(1)
	}

	lines := strings.Split(string(out), "\n")
	var audioOutputs []string
	for _, line := range lines {
		if line != "" {
			parts := strings.Split(line, "\t")
			if len(parts) > 1 {
				audioOutputs = append(audioOutputs, parts[1])
			}
		}
	}

	// Get Bluetooth Devices
	bluetoothDevices, err := getBluetoothDevices()
	if err != nil {
		fmt.Println("Warning: could not get bluetooth devices:", err)
	}

	// Get Airplane Mode status
	airplaneStatus, err := getAirplaneModeStatus()
	if err != nil {
		fmt.Println("Warning: could not get airplane mode status:", err)
	}

	p := tea.NewProgram(model{
		menuItems:           []string{"Audio Output", "Bluetooth", "Airplane Mode"}, // Updated
		audioOutputs:        audioOutputs,
		bluetoothDevices:    bluetoothDevices,
		airplaneModeEnabled: airplaneStatus, // Added
	})
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
