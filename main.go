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
	menuItems               []string
	audioOutputs            []string
	bluetoothDevices        []string
	wifiNetworks            []string // Added
	cursor                  int
	audioOutputsVisible     bool
	bluetoothDevicesVisible bool
	wifiNetworksVisible     bool // Added
	airplaneModeEnabled     bool
	quitting                bool
	wifiLoading             bool
}

type wifiNetworksMsg []string

func (m model) Init() tea.Cmd {
	return func() tea.Msg {
		networks, err := getWifiNetworks()
		if err != nil {
			// Handle the error appropriately, maybe log it or send an error message
			return nil
		}
		return wifiNetworksMsg(networks)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case wifiNetworksMsg:
		m.wifiNetworks = msg
		m.wifiLoading = false
		return m, nil

	case tea.KeyMsg:
		// --- Dynamic Index Calculation ---
		// Calculate the current index of each main menu item
		// based on which sub-menus are open.
		audioItemIndex := 0

		bluetoothItemIndex := 1
		if m.audioOutputsVisible {
			bluetoothItemIndex += len(m.audioOutputs)
		}

		wifiItemIndex := 2
		if m.audioOutputsVisible {
			wifiItemIndex += len(m.audioOutputs)
		}
		if m.bluetoothDevicesVisible {
			wifiItemIndex += len(m.bluetoothDevices)
		}

		airplaneModeItemIndex := 3
		if m.audioOutputsVisible {
			airplaneModeItemIndex += len(m.audioOutputs)
		}
		if m.bluetoothDevicesVisible {
			airplaneModeItemIndex += len(m.bluetoothDevices)
		}
		if m.wifiNetworksVisible {
			airplaneModeItemIndex += len(m.wifiNetworks)
		}
		// --- End Dynamic Index Calculation ---

		switch msg.String() {

		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "esc":
			if m.audioOutputsVisible {
				m.audioOutputsVisible = false
				m.cursor = 0 // Base index
			} else if m.bluetoothDevicesVisible {
				m.bluetoothDevicesVisible = false
				m.cursor = 1 // Base index
			} else if m.wifiNetworksVisible { // Added
				m.wifiNetworksVisible = false
				m.cursor = 2 // Base index
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
			if m.wifiNetworksVisible { // Added
				totalItems += len(m.wifiNetworks)
			}
			if m.cursor < totalItems-1 {
				m.cursor++
			}

		case "l": // Expand
			if m.cursor == audioItemIndex { // On "Audio Output"
				m.audioOutputsVisible = true
				m.bluetoothDevicesVisible = false
				m.wifiNetworksVisible = false // Collapse others
				if len(m.audioOutputs) > 0 {
					m.cursor = audioItemIndex + 1
				}
			} else if m.cursor == bluetoothItemIndex { // On "Bluetooth"
				m.audioOutputsVisible = false
				m.bluetoothDevicesVisible = true
				m.wifiNetworksVisible = false // Collapse others
				if len(m.bluetoothDevices) > 0 {
					m.cursor = bluetoothItemIndex + 1
				}
			} else if m.cursor == wifiItemIndex { // On "WiFi Network"
				m.audioOutputsVisible = false
				m.bluetoothDevicesVisible = false
				m.wifiNetworksVisible = true // Expand this
				if len(m.wifiNetworks) > 0 {
					m.cursor = wifiItemIndex + 1
				}
			}

		case "h": // Collapse
			if m.audioOutputsVisible && m.cursor >= audioItemIndex && m.cursor <= audioItemIndex+len(m.audioOutputs) {
				m.audioOutputsVisible = false
				m.cursor = 0 // Base index
			} else if m.bluetoothDevicesVisible && m.cursor >= bluetoothItemIndex && m.cursor <= bluetoothItemIndex+len(m.bluetoothDevices) {
				m.bluetoothDevicesVisible = false
				m.cursor = 1 // Base index
			} else if m.wifiNetworksVisible && m.cursor >= wifiItemIndex && m.cursor <= wifiItemIndex+len(m.wifiNetworks) { // Added
				m.wifiNetworksVisible = false
				m.cursor = 2 // Base index
			}

		case "enter", " ":
			if m.cursor == audioItemIndex { // "Audio Output"
				m.audioOutputsVisible = !m.audioOutputsVisible
				m.bluetoothDevicesVisible = false
				m.wifiNetworksVisible = false
				if m.audioOutputsVisible && len(m.audioOutputs) > 0 {
					m.cursor = audioItemIndex + 1
				} else {
					m.cursor = audioItemIndex
				}
			} else if m.cursor == bluetoothItemIndex { // "Bluetooth"
				m.audioOutputsVisible = false
				m.bluetoothDevicesVisible = !m.bluetoothDevicesVisible
				m.wifiNetworksVisible = false
				if m.bluetoothDevicesVisible && len(m.bluetoothDevices) > 0 {
					m.cursor = bluetoothItemIndex + 1
				} else {
					m.cursor = bluetoothItemIndex
				}
			} else if m.cursor == wifiItemIndex { // "WiFi Network"
				m.audioOutputsVisible = false
				m.bluetoothDevicesVisible = false
				m.wifiNetworksVisible = !m.wifiNetworksVisible
				if m.wifiNetworksVisible && len(m.wifiNetworks) > 0 {
					m.cursor = wifiItemIndex + 1
				} else {
					m.cursor = wifiItemIndex
				}
			} else if m.cursor == airplaneModeItemIndex { // "Airplane Mode"
				m.airplaneModeEnabled = !m.airplaneModeEnabled
				m.audioOutputsVisible = false
				m.bluetoothDevicesVisible = false
				m.wifiNetworksVisible = false // Collapse others

				var cmd *exec.Cmd
				if m.airplaneModeEnabled {
					cmd = exec.Command("nmcli", "radio", "all", "off")
				} else {
					cmd = exec.Command("nmcli", "radio", "all", "on")
				}
				cmd.Run()

			} else if m.audioOutputsVisible && m.cursor > audioItemIndex && m.cursor <= audioItemIndex+len(m.audioOutputs) {
				// An audio device is selected
				return m, tea.Quit
			} else if m.bluetoothDevicesVisible && m.cursor > bluetoothItemIndex && m.cursor <= bluetoothItemIndex+len(m.bluetoothDevices) {
				// A bluetooth device is selected
				return m, tea.Quit
			} else if m.wifiNetworksVisible && m.cursor > wifiItemIndex && m.cursor <= wifiItemIndex+len(m.wifiNetworks) { // Added
				// A wifi network is selected
				// Add connect logic here
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
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

		// Render WiFi networks if visible
		if item == "WiFi Network" && m.wifiNetworksVisible {
			if m.wifiLoading {
				s += "    Loading...\n"
			} else {
				for _, wifiNetwork := range m.wifiNetworks {
					wifiCursor := " "
					if m.cursor == currentItemIndex {
						wifiCursor = ">"
					}
					wifiLine := fmt.Sprintf("  %s %s", wifiCursor, wifiNetwork)
					if m.cursor == currentItemIndex {
						s += selectedStyle.Render(wifiLine) + "\n"
					} else {
						s += wifiLine + "\n"
					}
					currentItemIndex++
				}
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

// getWifiNetworks runs `nmcli dev wifi list` and parses SSIDs
func getWifiNetworks() ([]string, error) {
	// -t for terse (scriptable) output, -f for fields, rescan
	cmd := exec.Command("nmcli", "-t", "-f", "SSID", "dev", "wifi", "list", "--rescan", "yes")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("could not run nmcli: %w. Is NetworkManager running?", err)
	}

	lines := strings.Split(string(out), "\n")
	var networks []string
	seen := make(map[string]bool) // De-duplicate SSIDs

	for _, line := range lines {
		ssid := line
		// nmcli can list the same SSID multiple times for different bands/BSSIDs
		// or list empty lines.
		if ssid == "" || seen[ssid] {
			continue
		}
		networks = append(networks, ssid)
		seen[ssid] = true
	}
	return networks, nil
}

// getAirplaneModeStatus checks if 'nmcli radio wifi' reports 'disabled'
func getAirplaneModeStatus() (bool, error) {
	cmd := exec.Command("nmcli", "radio", "wifi")
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("could not check nmcli: %w. Is NetworkManager installed?", err)
	}

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
		menuItems:           []string{"Audio Output", "Bluetooth", "WiFi Network", "Airplane Mode"}, // Updated
		audioOutputs:        audioOutputs,
		bluetoothDevices:    bluetoothDevices,
		wifiNetworks:        []string{}, // Initialized as empty
		airplaneModeEnabled: airplaneStatus,
		wifiLoading:         true,
	})
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
