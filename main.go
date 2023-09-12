package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const localaiURL = "http://localhost:8080"
const localaiExecutable = "/Users/Aidan/Localai/local-ai"
var localaiProcess *exec.Cmd
var menuOptions = []string{"Chat", "Completion", "Exit"}
var selectedIndex int

func main() {
	// Start localai
	localaiProcess = exec.Command(localaiExecutable)
	if err := localaiProcess.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start localai: %v", err)
		return
	}
	time.Sleep(2 * time.Second)

	p := tea.NewProgram(model{})
	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v", err)
		os.Exit(1)
	}

	if err := localaiProcess.Process.Signal(syscall.SIGINT); err != nil {
		fmt.Fprintf(os.Stderr, "Error stopping localai: %v", err)
	}
}

type model struct{}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "s", "ArrowDown":
			selectedIndex = (selectedIndex + 1) % len(menuOptions)
		case "k", "w", "ArrowUp":
			selectedIndex = (selectedIndex - 1 + len(menuOptions)) % len(menuOptions)
		case "q", "Esc":
			return m, tea.Quit
		case "Enter", " ":
			switch menuOptions[selectedIndex] {
			case "Chat":
				fmt.Println(callChatAPI())
			case "Completion":
				fmt.Println(callCompletionAPI())
			case "Exit":
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder
	for i, option := range menuOptions {
		if i == selectedIndex {
			b.WriteString("> ")
		} else {
			b.WriteString("  ")
		}
		b.WriteString(option + "\n")
	}
	return b.String()
}

func callChatAPI() string {
	data := map[string]interface{}{
		"model":       "lunademo",
		"messages":    []map[string]string{{"role": "user", "content": "How are you?"}},
		"temperature": 0.9,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Sprintf("Error preparing data: %v", err)
	}

	resp, err := http.Post(localaiURL+"/v1/chat/completions", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Sprintf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Error reading response: %v", err)
	}

	return string(body)
}

func callCompletionAPI() string {
	data := map[string]interface{}{
		"model":       "lunademo",
		"prompt":      "function downloadFile(string url, string outputPath) {",
		"max_tokens":  256,
		"temperature": 0.5,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Sprintf("Error preparing data: %v", err)
	}

	resp, err := http.Post(localaiURL+"/v1/completions", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Sprintf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Error reading response: %v", err)
	}

	return string(body)
}

