# Quick Settings CLI
![Screencast From 2025-11-13 10-56-36](https://github.com/user-attachments/assets/2b44c33f-52ff-4127-a89a-9134e2745cf9)

A command-line tool for managing system settings like audio output, Bluetooth, Wi-Fi, and airplane mode. This tool provides a simple and efficient way to manage your system's settings directly from the terminal.

## Features

- **Audio Output:** View and select audio output devices.
- **Bluetooth:** View and manage Bluetooth devices.
- **Wi-Fi:** View and connect to Wi-Fi networks.
- **Airplane Mode:** Toggle airplane mode on and off.

## Dependencies

This project is built with Go and uses the following key libraries:

- [Bubble Tea](https://github.com/charmbracelet/bubbletea): A powerful framework for building terminal-based user interfaces.
- [Lip Gloss](https://github.com/charmbracelet/lipgloss): A library for styling terminal output.

This tool also relies on the following system commands:

- `pactl`: For managing audio devices.
- `bluetoothctl`: For managing Bluetooth devices.
- `nmcli`: For managing network connections.

## Installation

1. **Clone the repository:**
   ```sh
   git clone https://github.com/your-username/quick-settings-cli.git
   cd quick-settings-cli
   ```

2. **Build the project:**
   ```sh
   go build
   ```

3. **Run the application:**
   ```sh
   ./quick-settings-cli
   ```

## Usage

- Use the **up/down arrow keys** or **j/k** to navigate the menu.
- Press **l** or **enter** to expand a menu item.
- Press **h** or **esc** to collapse a menu item.
- Press **q** or **ctrl+c** to quit the application.

## TODO

- [ ] Add support for connecting to Bluetooth devices.
- [ ] Add support for connecting to Wi-Fi networks.
- [ ] Add support for more system settings (e.g., display brightness, volume control).
- [ ] Add a configuration file for customizing the tool's behavior.
- [ ] Improve error handling and provide more informative error messages.
