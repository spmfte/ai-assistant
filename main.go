package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/karrick/godirwalk"
)

type model struct {
	state appState
}

type appState int

const (
	mainMenu appState = iota
	textMenu
	imageMenu
	audioMenu
	fileBrowser
	displayResult
)

var (
	menuOptions       = []string{"Text", "Image", "Audio", "Exit"}
	selectedMenuIndex int
	currentDirectory  = "/Users/aidan/ai-assistant/mydata"
	files             []string
	selectedFileIndex int
	analysisResult    string
)

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case mainMenu:
			switch msg.String() {
			case "j", "ArrowDown":
				selectedMenuIndex = (selectedMenuIndex + 1) % len(menuOptions)
			case "k", "ArrowUp":
				selectedMenuIndex = (selectedMenuIndex - 1 + len(menuOptions)) % len(menuOptions)
			case "q", "Esc":
				return m, tea.Quit
			case "enter", "\r":
				switch selectedMenuIndex {
				case 0, 1, 2:
					files, err := listFiles(currentDirectory)
					if err != nil {
						fmt.Println("Error listing files:", err)
						return m, nil
					}
					if len(files) > 0 {
						m.state = fileBrowser
					} else {
						return m, nil
					}
				case 3:
					return m, tea.Quit
				}
			}
		case fileBrowser:
			if len(files) == 0 {
				return m, nil
			}
			switch msg.String() {
			case "j", "ArrowDown":
				selectedFileIndex = (selectedFileIndex + 1) % len(files)
			case "k", "ArrowUp":
				selectedFileIndex = (selectedFileIndex - 1 + len(files)) % len(files)
      case "enter", "\r":
      	var err error
	      analysisResult, err = sendFileToAI(files[selectedFileIndex])
	      if err != nil {
	      	fmt.Println("Error sending file to AI:", err)
		      return m, nil
	}     else {
		m.state = displayResult
	}
					case "q", "Esc":
				m.state = mainMenu
			}
		case displayResult:
			switch msg.String() {
			case "q", "Esc":
				m.state = mainMenu
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	switch m.state {
	case mainMenu, textMenu, imageMenu, audioMenu:
		b.WriteString("Select an Option:\n\n")
		for i, option := range menuOptions {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFF"))
			if i == selectedMenuIndex {
				style = style.Background(lipgloss.Color("#357DED")).Foreground(lipgloss.Color("#FFF")).Bold(true)
			}
			b.WriteString(style.Render(option) + "\n")
		}
	case fileBrowser:
		if len(files) == 0 {
			b.WriteString("No files found.\n")
		} else {
			b.WriteString("Select a File:\n\n")
			for i, file := range files {
				style := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFF"))
				if i == selectedFileIndex {
					style = style.Background(lipgloss.Color("#357DED")).Foreground(lipgloss.Color("#FFF")).Bold(true)
				}
				b.WriteString(style.Render(file) + "\n")
			}
		}
	case displayResult:
		b.WriteString("AI Analysis Result:\n")
		b.WriteString(analysisResult)
	}

	return b.String()
}

func main() {
	p := tea.NewProgram(model{state: mainMenu})
	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v", err)
		os.Exit(1)
	}
}

func sendFileToAI(filePath string) (string, error) {
	url := "http://localai.server.endpoint/analyze"

	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, file)

	err = writer.Close()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(respBody), nil
}

func listFiles(directory string) ([]string, error) {
	var filesList []string
	err := godirwalk.Walk(directory, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if !de.IsDir() {
				filesList = append(filesList, osPathname)
			}
			return nil
		},
		Unsorted: true,
	})
	if err != nil {
		return nil, err
	}
	return filesList, nil
}

