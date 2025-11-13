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
	menuItems           []string
	audioOutputs        []string
	cursor              int
	audioOutputsVisible bool
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
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			totalItems := len(m.menuItems)
			if m.audioOutputsVisible {
				totalItems += len(m.audioOutputs)
			}
			if m.cursor < totalItems-1 {
				m.cursor++
			}
		case "l":
			if m.cursor == 0 { // "Audio Output" is selected
				m.audioOutputsVisible = true
				m.cursor = 1 // Move cursor to the first audio output
			}
		case "h":
			if m.audioOutputsVisible {
				m.audioOutputsVisible = false
				m.cursor = 0
			}
		case "enter", " ":
			if m.cursor == 0 { // "Audio Output" is selected
				m.audioOutputsVisible = !m.audioOutputsVisible
				if m.audioOutputsVisible {
					m.cursor = 1 // Move cursor to the first audio output if revealing
				}
			} else if m.audioOutputsVisible && m.cursor > 0 {
				// An audio device is selected
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	s := "Quick Settings\n\n"

	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))

	// Render main menu
	for i, item := range m.menuItems {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		line := fmt.Sprintf("%s %s", cursor, item)
		if m.cursor == i {
			s += selectedStyle.Render(line) + "\n"
		} else {
			s += line + "\n"
		}

		// Render audio outputs if visible
		if item == "Audio Output" && m.audioOutputsVisible {
			for j, audioDevice := range m.audioOutputs {
				audioCursor := " "
				audioIndex := i + j + 1
				if m.cursor == audioIndex {
					audioCursor = ">"
				}
				audioLine := fmt.Sprintf("  %s %s", audioCursor, audioDevice)
				if m.cursor == audioIndex {
					s += selectedStyle.Render(audioLine) + "\n"
				} else {
					s += audioLine + "\n"
				}
			}
		}
	}

	s += "\nPress l to expand, h to collapse, enter to select, q to quit.\n"
	return s
}

func main() {
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
			audioOutputs = append(audioOutputs, parts[1])
		}
	}

	p := tea.NewProgram(model{
		menuItems:    []string{"Audio Output"},
		audioOutputs: audioOutputs,
	})
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

