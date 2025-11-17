package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
	"os"
	"os/exec"
	"quick-settings-cli/bluetooth"
	"strings"
	"time"
)

type model struct {
	menuItems               []string
	audioOutputs            []Pair
	bluetoothDevices        []Pair
	wifiNetworks            []Pair
	cursor                  int
	audioOutputsVisible     bool
	bluetoothDevicesVisible bool
	wifiNetworksVisible     bool // Added
	airplaneModeEnabled     bool
	quitting                bool
	wifiLoading             bool
	connectingWifiNetwork   string
}

type Pair struct {
	name      string
	connected bool
}

type wifiNetworksMsg []Pair
type audioOutputsMsg []Pair
type bluetoothDevicesMsg []Pair

type tickMsg struct{}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		fetchWifiNetworksCmd(),
		fetchAudioOutputsCmd(),
		fetchBluetoothDevicesCmd(),
		doTick(),
	)
}

func fetchWifiNetworksCmd() tea.Cmd {
	return func() tea.Msg {
		networks, err := getWifiNetworks()
		if err != nil {
			// Handle the error appropriately
			return nil
		}
		return wifiNetworksMsg(networks)
	}
}

func fetchAudioOutputsCmd() tea.Cmd {
	return func() tea.Msg {
		audioOutputs, err := getAudioOutputs()
		if err != nil {
			// Handle the error appropriately
			return nil
		}
		return audioOutputsMsg(audioOutputs)
	}
}

func fetchBluetoothDevicesCmd() tea.Cmd {
	return func() tea.Msg {
		bluetoothDevices, err := getBluetoothDevices()
		if err != nil {
			// Handle the error appropriately
			return nil
		}
		return bluetoothDevicesMsg(bluetoothDevices)
	}
}

func doTick() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case wifiNetworksMsg:
		m.wifiNetworks = msg
		m.wifiLoading = false
		m.connectingWifiNetwork = ""
		m.clampCursor()
		return m, nil
	case audioOutputsMsg:
		m.audioOutputs = msg
		m.clampCursor()
		return m, nil

	case bluetoothDevicesMsg:
		m.bluetoothDevices = msg
		m.clampCursor()
		return m, nil

	case tickMsg:
		return m, tea.Batch(fetchWifiNetworksCmd(), fetchBluetoothDevicesCmd(), doTick())

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
				m.quitting = true
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
			switch m.cursor {
			case audioItemIndex: // On "Audio Output"
				m.audioOutputsVisible = true
				m.bluetoothDevicesVisible = false
				m.wifiNetworksVisible = false // Collapse others
				if len(m.audioOutputs) > 0 {
					m.cursor = audioItemIndex + 1
				}
			case bluetoothItemIndex: // On "Bluetooth"
				m.audioOutputsVisible = false
				m.bluetoothDevicesVisible = true
				m.wifiNetworksVisible = false // Collapse others
				if len(m.bluetoothDevices) > 0 {
					m.cursor = bluetoothItemIndex + 1
				}
			case wifiItemIndex: // On "WiFi Network"
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
				selectedIndex := m.cursor - (audioItemIndex + 1)
				if selectedIndex >= 0 && selectedIndex < len(m.audioOutputs) {
					selectedOutput := m.audioOutputs[selectedIndex]
					_ = selectAudioOutput(selectedOutput.name) // Handle error if necessary
					return m, fetchAudioOutputsCmd()
				}
				return m, nil
			} else if m.bluetoothDevicesVisible && m.cursor > bluetoothItemIndex && m.cursor <= bluetoothItemIndex+len(m.bluetoothDevices) {
				// A bluetooth device is selected
				selectedIndex := m.cursor - (bluetoothItemIndex + 1)
				if selectedIndex >= 0 && selectedIndex < len(m.bluetoothDevices) {
					selectedDevice := m.bluetoothDevices[selectedIndex]
					_ = selectBluetoothDevice(selectedDevice.name) // Handle error if necessary
					return m, fetchBluetoothDevicesCmd()
				}
				return m, nil
			} else if m.wifiNetworksVisible && m.cursor > wifiItemIndex && m.cursor <= wifiItemIndex+len(m.wifiNetworks) { // Added
				// A wifi network is selected
				selectedIndex := m.cursor - (wifiItemIndex + 1)
				if selectedIndex >= 0 && selectedIndex < len(m.wifiNetworks) {
					selectedNetwork := m.wifiNetworks[selectedIndex]
					m.connectingWifiNetwork = selectedNetwork.name // Set the network being connected
					m.wifiLoading = true                           // Set loading to true
					_ = selectWifiNetwork(selectedNetwork.name)    // Handle error if necessary
					return m, fetchWifiNetworksCmd()
				}
				return m, nil
			}
		}
	}
	return m, nil
}

func (m *model) clampCursor() {
	totalItems := len(m.menuItems)
	if m.audioOutputsVisible {
		totalItems += len(m.audioOutputs)
	}
	if m.bluetoothDevicesVisible {
		totalItems += len(m.bluetoothDevices)
	}
	if m.wifiNetworksVisible {
		totalItems += len(m.wifiNetworks)
	}
	if m.cursor >= totalItems {
		m.cursor = totalItems - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	s := "Quick Settings\n\n"

	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	statusStyle := lipgloss.NewStyle().Faint(true)

	const nameColumnWidth = 25 // Adjust as needed for proper alignment

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
			// Pad the item name to align the status
			paddedItem := fmt.Sprintf("%-*s", nameColumnWidth, item)
			line = fmt.Sprintf("%s %s %s", cursor, paddedItem, statusStyle.Render(status))
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
				status := "Disconnected"
				if audioDevice.connected {
					status = "Connected"
				}
				// Pad the device name to align the status
				paddedName := fmt.Sprintf("%-*s", nameColumnWidth, audioDevice.name)
				audioLine := fmt.Sprintf("  %s %s %s", audioCursor, paddedName, statusStyle.Render(status))
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
				status := "Disconnected"
				if btDevice.connected {
					status = "Connected"
				}
				// Pad the device name to align the status
				paddedName := fmt.Sprintf("%-*s", nameColumnWidth, btDevice.name)
				btLine := fmt.Sprintf("  %s %s %s", btCursor, paddedName, statusStyle.Render(status))
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

					status := "Disconnected"

					if wifiNetwork.connected {

						status = "Connected"

					}

					// If this is the network being connected, show 'Connecting...'

					if m.connectingWifiNetwork == wifiNetwork.name && m.wifiLoading {

						status = "Connecting..."

					}

					// Pad the network name to align the status

					paddedName := fmt.Sprintf("%-*s", nameColumnWidth, wifiNetwork.name)

					wifiLine := fmt.Sprintf("  %s %s %s", wifiCursor, paddedName, statusStyle.Render(status))

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
func getBluetoothDevices() ([]Pair, error) {
	cmd := exec.Command("bluetoothctl", "devices")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("could not run bluetoothctl: %w. Is bluetooth service running?", err)
	}

	lines := strings.Split(string(out), "\n")
	var devices []Pair
	for _, line := range lines {
		if line != "" {
			parts := strings.Split(line, " ")
			if len(parts) > 2 {
				deviceName := strings.Join(parts[2:], " ")
				addr := parts[1] // Extracting MAC address
				cmd := exec.Command("bluetoothctl", "info", addr)
				out, err := cmd.Output()
				if err != nil {
					return nil, fmt.Errorf("could not run bluetoothctl: %w. Is bluetooth service running?", err)
				}
				isPaired, err := bluetooth.ParsePairedStatus(string(out))
				if err != nil {
					return nil, fmt.Errorf("could not parse device informations correctly")
				}
				devices = append(devices, Pair{deviceName, bool(isPaired)})
			}
		}
	}
	return devices, nil
}

// selectBluetoothDevice connects to a Bluetooth device by its name
func selectBluetoothDevice(deviceName string) error {
	// First, find the MAC address of the device
	cmd := exec.Command("bluetoothctl", "devices")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("could not run bluetoothctl to list devices: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	var macAddress string
	for _, line := range lines {
		if strings.Contains(line, deviceName) {
			parts := strings.Split(line, " ")
			if len(parts) > 1 {
				macAddress = parts[1]
				break
			}
		}
	}

	if macAddress == "" {
		return fmt.Errorf("bluetooth device %s not found", deviceName)
	}

	// Connect to the device
	cmd = exec.Command("bluetoothctl", "connect", macAddress)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("could not connect to bluetooth device %s: %w", deviceName, err)
	}
	return nil
}

// getWifiNetworks runs `nmcli dev wifi list` and parses SSIDs
func getWifiNetworks() ([]Pair, error) {
	// Get currently active Wi-Fi network
	activeWifiCmd := exec.Command("nmcli", "-t", "-f", "active,ssid", "dev", "wifi")
	activeWifiOut, err := activeWifiCmd.Output()
	var activeSSID string
	if err == nil {
		activeWifiLines := strings.Split(string(activeWifiOut), "\n")
		for _, line := range activeWifiLines {
			if strings.HasPrefix(line, "yes:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					activeSSID = parts[1]
					break
				}
			}
		}
	}

	// -t for terse (scriptable) output, -f for fields, rescan
	cmd := exec.Command("nmcli", "-t", "-f", "SSID,ACTIVE", "dev", "wifi", "list", "--rescan", "yes")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("could not run nmcli: %w. Is NetworkManager running?", err)
	}

	lines := strings.Split(string(out), "\n")
	var networks []Pair
	seen := make(map[string]bool) // De-duplicate SSIDs

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}
		ssid := parts[0]
		active := parts[1] == "yes"

		// If this SSID matches the activeSSID, mark it as connected
		if ssid == activeSSID {
			active = true
		}

		if ssid == "" || seen[ssid] {
			continue
		}
		networks = append(networks, Pair{ssid, active})
		seen[ssid] = true
	}
	return networks, nil
}

// selectWifiNetwork connects to a WiFi network by its name
func selectWifiNetwork(networkName string) error {
	cmd := exec.Command("nmcli", "dev", "wifi", "connect", networkName)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("could not connect to wifi network %s: %w", networkName, err)
	}
	return nil
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

func getAudioOutputs() ([]Pair, error) {
	// Get Audio Devices
	cmd := exec.Command("pactl", "list", "short", "sinks")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("could not get audio devices:", err)
		os.Exit(1)
	}

	lines := strings.Split(string(out), "\n")
	var audioOutputs []Pair
	for _, line := range lines {
		if line != "" {
			parts := strings.Split(line, "\t")
			if len(parts) > 1 {
				outputName := parts[1]
				addr := parts[4] // Status
				isPaired := false
				if addr == "RUNNING" {
					isPaired = true
				}
				audioOutputs = append(audioOutputs, Pair{outputName, isPaired})
			}
		}
	}
	return audioOutputs, nil
}

// selectAudioOutput sets the default audio sink using pactl
func selectAudioOutput(outputName string) error {
	cmd := exec.Command("pactl", "set-default-sink", outputName)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("could not set default sink: %w", err)
	}
	return nil
}

func main() {
	// Get Audio Outputs
	audioOutputs, err := getAudioOutputs()
	if err != nil {
		fmt.Println("Warning: could not get audio outputs", err)
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
		wifiNetworks:        []Pair{}, // Initialized as empty
		airplaneModeEnabled: airplaneStatus,
		wifiLoading:         true,
	})
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
