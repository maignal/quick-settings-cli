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
	bluetoothDevices      []string // Added
	cursor                int
	audioOutputsVisible   bool
	bluetoothDevicesVisible bool // Added
}

func (m model) Init() tea.Cmd {
	// Init is the first function that will be called. It returns a command.
	// We don't need to do anything here, so we'll return nil.
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		
		// MODIFIED: q and ctrl+c always quit
		case "q", "ctrl+c":
			return m, tea.Quit

		// NEW: 'esc' conditionally collapses or quits
		case "esc":
			if m.audioOutputsVisible {
				m.audioOutputsVisible = false
				m.cursor = 0 // "Audio Output" is at index 0
			} else if m.bluetoothDevicesVisible {
				m.bluetoothDevicesVisible = false
				m.cursor = 1 // "Bluetooth" is at index 1
			} else {
				// No sub-menus open, so quit
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
				// Adjust index if audio list is expanded
				bluetoothItemIndex += len(m.audioOutputs)
			}

			if m.cursor == audioItemIndex { // On "Audio Output"
				m.audioOutputsVisible = true
				m.bluetoothDevicesVisible = false // Collapse other
				if len(m.audioOutputs) > 0 {
					m.cursor = audioItemIndex + 1 // Move to first audio item
				}
			} else if m.cursor == bluetoothItemIndex { // On "Bluetooth"
				m.bluetoothDevicesVisible = true
				m.audioOutputsVisible = false // Collapse other
				if len(m.bluetoothDevices) > 0 {
					m.cursor = bluetoothItemIndex + 1 // Move to first bluetooth item
				}
			}

		case "h": // Collapse
			audioItemIndex := 0
			bluetoothItemIndex := 1 // Base index in menuItems
			if m.audioOutputsVisible {
				// Adjust index if audio list is expanded
				bluetoothItemIndex += len(m.audioOutputs)
			}

			// If cursor is inside audio list
			if m.audioOutputsVisible && m.cursor > audioItemIndex && m.cursor <= audioItemIndex+len(m.audioOutputs) {
				m.audioOutputsVisible = false
				m.cursor = audioItemIndex // Move to "Audio Output"
			} else if m.bluetoothDevicesVisible && m.cursor > bluetoothItemIndex && m.cursor <= bluetoothItemIndex+len(m.bluetoothDevices) {
				// If cursor is inside bluetooth list
				m.bluetoothDevicesVisible = false
				// After collapse, bluetooth index is just 1
				m.cursor = 1 // Move to "Bluetooth"
			}

		case "enter", " ":
			audioItemIndex := 0
			bluetoothItemIndex := 1 // Base index in menuItems
			if m.audioOutputsVisible {
				// Adjust index if audio list is expanded
				bluetoothItemIndex += len(m.audioOutputs)
			}

			if m.cursor == audioItemIndex { // "Audio Output" is selected
				m.audioOutputsVisible = !m.audioOutputsVisible
				m.bluetoothDevicesVisible = false // Collapse other
				if m.audioOutputsVisible && len(m.audioOutputs) > 0 {
					m.cursor = audioItemIndex + 1 // Move to first audio item
				} else {
					m.cursor = audioItemIndex // Stay on parent if closing
				}
			} else if m.cursor == bluetoothItemIndex { // "Bluetooth" is selected
				m.bluetoothDevicesVisible = !m.bluetoothDevicesVisible
				m.audioOutputsVisible = false // Collapse other
				if m.bluetoothDevicesVisible && len(m.bluetoothDevices) > 0 {
					m.cursor = bluetoothItemIndex + 1 // Move to first bluetooth item
				} else {
					m.cursor = bluetoothItemIndex // Stay on parent if closing
				}
			} else if m.audioOutputsVisible && m.cursor > audioItemIndex && m.cursor <= audioItemIndex+len(m.audioOutputs) {
				// An audio device is selected
				// You could add logic here to set the device
				return m, tea.Quit
			} else if m.bluetoothDevicesVisible && m.cursor > bluetoothItemIndex && m.cursor <= bluetoothItemIndex+len(m.bluetoothDevices) {
				// A bluetooth device is selected
				// You could add logic here to connect to the device
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	s := "Quick Settings\n\n"

	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))

	// currentItemIndex tracks the absolute cursor position
	currentItemIndex := 0

	// Render main menu
	for _, item := range m.menuItems {
		cursor := " "
		if m.cursor == currentItemIndex {
			cursor = ">"
		}
		line := fmt.Sprintf("%s %s", cursor, item)
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
				currentItemIndex++ // Increment for the sub-item
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
				currentItemIndex++ // Increment for the sub-item
			}
		}
	}

	s += "\nPress l to expand, h to collapse, enter to select, q to quit.\n"
	return s
}

// getBluetoothDevices runs `bluetoothctl devices` and parses the output
func getBluetoothDevices() ([]string, error) {
	cmd := exec.Command("bluetoothctl", "devices")
	out, err := cmd.Output()
	if err != nil {
		// Note: bluetoothctl might not be installed or service not running
		return nil, fmt.Errorf("could not run bluetoothctl: %w. Is bluetooth service running?", err)
	}

	lines := strings.Split(string(out), "\n")
	var devices []string
	for _, line := range lines {
		if line != "" {
			parts := strings.Split(line, " ")
			if len(parts) > 2 {
				// Output is "Device XX:XX:XX:XX:XX:XX Device Name"
				deviceName := strings.Join(parts[2:], " ")
				devices = append(devices, deviceName)
			}
		}
	}
	return devices, nil
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
		// Don't exit, just show a warning and continue with an empty list
		fmt.Println("Warning: could not get bluetooth devices:", err)
	}

	p := tea.NewProgram(model{
		menuItems:        []string{"Audio Output", "Bluetooth"}, // Updated
		audioOutputs:     audioOutputs,
		bluetoothDevices: bluetoothDevices, // Added
	})
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
