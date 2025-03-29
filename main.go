/*
File Concatenator CLI

This CLI tool allows you to concatenate specific file types from a directory into a single output file.
It provides both a wizard-style interface and command-line options for flexibility.

Usage:
  filecon [flags]

Flags:
  -d, --dir string       Directory to search for files (default is current directory)
  -e, --ext string       File extension to search for (e.g., .go, .js, .py)
  -o, --out string       Output file name (default is "filecon_output.txt")
  -r, --remove-spaces    Remove all tabs and extra spaces from the content (default false)
  -i, --ignore-ext       File extensions to ignore (e.g., .tmp, .bak)
  -f, --ignore-folders   Folders to ignore during search (e.g., node_modules, .git)
  -h, --help             Help for filecon

Examples:
  1. Run the interactive wizard:
     filecon

  2. Concatenate all .go files in the current directory into filecon_output.txt:
     filecon --dir=. --ext=.go --out=output.txt

  3. Concatenate all .js files in /path/to/dir into filecon_result.js, removing extra spaces:
     filecon --dir=/path/to/dir --ext=.js --out=result.js --remove-spaces

  4. Concatenate all .py files in the current directory, ignoring .pyc files and the venv folder:
     filecon --dir=. --ext=.py --ignore-ext=.pyc --ignore-folders=venv,__pycache__

Note: If you don't provide all required flags (dir, ext, out), the interactive wizard will start.
*/

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

const (
	version = "1.1.0"
	gitRepo = "https://github.com/filecon/filecon"
)

var (
	dir           string
	fileType      string
	outputFile    string
	removeSpaces  bool
	ignoreExts    string
	ignoreFolders string
)

var rootCmd = &cobra.Command{
	Use:   "filecon",
	Short: "A CLI tool to concatenate specific file types",
	Long:  `File Concatenator is a CLI application that allows you to search a directory for specific file types and concatenate their content into a single output file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if dir == "" {
			dir = "."
		}
		if dir == "/" {
			fmt.Print("Warning: You're about to concatenate files from the root directory. Are you sure? (y/N): ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" {
				fmt.Println("Operation cancelled.")
				return
			}
		}
		if fileType != "" && outputFile != "" {
			// Add filecon_ prefix to the output file if not already there
			if !strings.HasPrefix(outputFile, "filecon_") {
				outputFile = "filecon_" + outputFile
			}

			// Process the ignore extensions and folders
			ignoreExtList := strings.Split(ignoreExts, ",")
			ignoreFolderList := strings.Split(ignoreFolders, ",")

			if err := concatenateFiles(dir, fileType, outputFile, removeSpaces, ignoreExtList, ignoreFolderList); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("File concatenation completed successfully!")
		} else {
			p := tea.NewProgram(initialModel())
			if _, err := p.Run(); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.Flags().StringVarP(&dir, "dir", "d", "", "Directory to search for files (default is current directory)")
	rootCmd.Flags().StringVarP(&fileType, "ext", "e", "", "File extension to search for")
	rootCmd.Flags().StringVarP(&outputFile, "out", "o", "", "Output file name (optional)")
	rootCmd.Flags().BoolVarP(&removeSpaces, "remove-spaces", "r", false, "Remove all tabs and extra spaces from the content")
	rootCmd.Flags().StringVarP(&ignoreExts, "ignore-ext", "i", "", "Comma-separated list of file extensions to ignore")
	rootCmd.Flags().StringVarP(&ignoreFolders, "ignore-folders", "f", "", "Comma-separated list of folders to ignore")
}

type model struct {
	inputs        []textinput.Model
	currentInput  int
	err           error
	done          bool
	removeSpaces  bool
	ignoreExts    string
	ignoreFolders string
}

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle.Copy()
	noStyle      = lipgloss.NewStyle()

	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

func initialModel() model {
	m := model{
		inputs: make([]textinput.Model, 5),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.CursorStyle = cursorStyle
		t.CharLimit = 256

		switch i {
		case 0:
			t.Placeholder = "Directory (e.g., ., default is current directory)"
			t.Focus()
		case 1:
			t.Placeholder = "File extension (e.g., .dart)"
		case 2:
			t.Placeholder = "Output file (optional, default is filecon_<timestamp>.txt)"
		case 3:
			t.Placeholder = "Extensions to ignore (comma-separated, e.g., .tmp,.bak)"
		case 4:
			t.Placeholder = "Folders to ignore (comma-separated, e.g., node_modules,.git)"
		}

		m.inputs[i] = t
	}

	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.currentInput == len(m.inputs) {
				dir := m.inputs[0].Value()
				if dir == "" {
					dir = "."
				}
				if dir == "/" {
					fmt.Print("Warning: You're about to concatenate files from the root directory. Are you sure? (y/N): ")
					var response string
					fmt.Scanln(&response)
					if strings.ToLower(response) != "y" {
						fmt.Println("Operation cancelled.")
						return m, tea.Quit
					}
				}
				outputFile := m.inputs[2].Value()
				if outputFile == "" {
					outputFile = fmt.Sprintf("filecon_%s.txt", time.Now().Format("20060102_150405"))
				} else if !strings.HasPrefix(outputFile, "filecon_") {
					outputFile = "filecon_" + outputFile
				}

				ignoreExtList := strings.Split(m.inputs[3].Value(), ",")
				ignoreFolderList := strings.Split(m.inputs[4].Value(), ",")

				m.err = concatenateFiles(dir, m.inputs[1].Value(), outputFile, m.removeSpaces, ignoreExtList, ignoreFolderList)
				m.done = true
				return m, tea.Quit
			}

			if s == "up" || s == "shift+tab" {
				m.currentInput--
			} else {
				m.currentInput++
			}

			if m.currentInput > len(m.inputs) {
				m.currentInput = 0
			} else if m.currentInput < 0 {
				m.currentInput = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.currentInput {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		case " ":
			if m.currentInput == len(m.inputs) {
				m.removeSpaces = !m.removeSpaces
			}
			return m, nil
		}
	}

	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress any key to exit.", m.err)
	}
	if m.done {
		return "File concatenation completed successfully!\nPress any key to exit."
	}

	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	removeSpacesStatus := "[ ] Remove extra spaces"
	if m.removeSpaces {
		removeSpacesStatus = "[x] Remove extra spaces"
	}

	button := &blurredButton
	removeSpacesStyle := blurredStyle
	if m.currentInput == len(m.inputs) {
		button = &focusedButton
		removeSpacesStyle = focusedStyle
	}

	fmt.Fprintf(&b, "\n\n%s\n", removeSpacesStyle.Render(removeSpacesStatus))
	fmt.Fprintf(&b, "\n%s\n\n", *button)

	b.WriteString(helpStyle.Render("cursor: ↑↓ • toggle option: space • submit: enter • quit: esc"))

	return b.String()
}

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func concatenateFiles(dir, fileType, outputFile string, removeSpaces bool, ignoreExts, ignoreFolders []string) error {
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer outFile.Close()

	// Write the filecon signature at the top of the file
	signature := fmt.Sprintf("# Generated by File Concatenator (filecon) v%s\n# %s\n# Generated on: %s\n\n",
		version,
		gitRepo,
		time.Now().Format("2006-01-02 15:04:05"))

	if _, err = outFile.WriteString(signature); err != nil {
		return fmt.Errorf("error writing signature to output file: %v", err)
	}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories that match any in the ignore list
		if info.IsDir() {
			for _, folder := range ignoreFolders {
				folder = strings.TrimSpace(folder)
				if folder != "" && strings.HasSuffix(path, folder) {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Skip files with extensions in the ignore list
		for _, ext := range ignoreExts {
			ext = strings.TrimSpace(ext)
			if ext != "" && strings.HasSuffix(info.Name(), ext) {
				return nil
			}
		}

		// Process files with the target extension
		if strings.HasSuffix(info.Name(), fileType) {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("error reading file %s: %v", path, err)
			}

			if removeSpaces {
				content = removeTabsAndSpaces(content)
			}

			if _, err = outFile.WriteString(fmt.Sprintf("# %s\n---\n", path)); err != nil {
				return fmt.Errorf("error writing to output file: %v", err)
			}
			if _, err = outFile.WriteString(string(content) + "\n\n"); err != nil {
				return fmt.Errorf("error writing file content to output file: %v", err)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking the directory: %v", err)
	}
	return nil
}

func removeTabsAndSpaces(content []byte) []byte {
	// Remove tabs
	content = bytes.ReplaceAll(content, []byte("\t"), []byte(""))

	// Remove extra spaces
	re := regexp.MustCompile(`\s+`)
	content = re.ReplaceAll(content, []byte(" "))

	// Trim leading and trailing spaces from each line
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return []byte(strings.Join(lines, "\n"))
}
