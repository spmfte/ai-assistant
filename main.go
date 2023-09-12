package main

import (
	"fmt"
	"os"
	"strings"
   tea "github.com/charmbracelet/bubbletea"
  "github.com/charmbracelet/lipgloss"
	"github.com/karrick/godirwalk"
  "net/http"
  "mime/multipart"
  "bytes"
  "io/ioutil"
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
	currentDirectory  = "/"
	files             []string
	selectedFileIndex int
	analysisResult    string
)

 func (m model) Init() tea.Cmd{
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch m.state {
        case mainMenu, textMenu, imageMenu, audioMenu:
            switch msg.String() {
            case "j", "ArrowDown":
                selectedMenuIndex = (selectedMenuIndex + 1) % len(menuOptions)
            case "k", "ArrowUp":
                selectedMenuIndex = (selectedMenuIndex - 1 + len(menuOptions)) % len(menuOptions)
            case "q", "Esc":
                return m, tea.Quit
            case "Enter":
                switch selectedMenuIndex {
                case 0, 1, 2:
                    m.state = fileBrowser
                    var err error
                    files, err = listFiles(currentDirectory)
                    if err != nil {
                        // Handle the error, maybe change the state to an error state
                        fmt.Println("Error listing files:", err)
                    }
                case 3:
                    return m, tea.Quit
                }
            }
       case fileBrowser:
    switch msg.String() {
    case "j", "ArrowDown":
        selectedFileIndex = (selectedFileIndex + 1) % len(files)
    case "k", "ArrowUp":
        selectedFileIndex = (selectedFileIndex - 1 + len(files)) % len(files)
    case "Enter":
        analysisResult, err := sendFileToAI(files[selectedFileIndex])
        if err != nil {
            // Handle the error. Here, I just print it, but you may want to 
            // change the state to an error state or do something else.
            fmt.Println("Error sending file to AI:", err)
        } else {
            m.state = displayResult
        }
    case "q", "Esc":
        m.state = mainMenu
    }            }
        case displayResult:
            // Handle displaying results logic here
        }
    }
    return m, nil
}

//... (rest of the code below this point)

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
		for i, file := range files {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFF"))
			if i == selectedFileIndex {
				style = style.Background(lipgloss.Color("#357DED")).Foreground(lipgloss.Color("#FFF")).Bold(true)
			}
			b.WriteString(style.Render(file) + "\n")
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

// ... Other functions remain unchanged ...

// Dummy function to represent sending a file to LocalAI
func sendFileToAI(filePath string) (string, error) {
    url := "http://localai.server.endpoint/analyze" // Replace this with the actual endpoint of your LocalAI server

    // Prepare the file for upload
    file, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer file.Close()

    // Create a buffer to store our request
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile("file", filepath.Base(filePath))
    if err != nil {
        return "", err
    }
    _, err = io.Copy(part, file)

    // Important: Close the writer to finish sending the file
    err = writer.Close()
    if err != nil {
        return "", err
    }

    // Create a new POST request to the server
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

    // Read the response from the server
    respBody, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    // Here, I am assuming that the response is a string. 
    // Adjust based on the actual response format (e.g., JSON).
    return string(respBody), nil
}
func listFiles(directory string) ([]string, error) {
	var files []string
	err := godirwalk.Walk(directory, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if !de.IsDir() {
				files = append(files, osPathname)
			}
			return nil
		},
		Unsorted: true, // Set true for faster file listing
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

