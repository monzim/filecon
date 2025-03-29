/*
File Concatenator CLI

This CLI tool allows you to concatenate specific file types from a directory into a single output file.
It provides both a wizard-style interface and command-line options for flexibility.

Usage:
  filecon [flags]

Flags:
  -d, --dir string       Directory to search for files (default is current directory)
  -e, --ext string       File extension to search for (e.g., go, js, py - with or without dot)
  -o, --out string       Output file name (default is "filecon_output.txt")
  -r, --remove-spaces    Remove all tabs and extra spaces from the content (default false)
  -i, --ignore-ext       File extensions to ignore (e.g., tmp, bak - with or without dot)
  -f, --ignore-folders   Folders to ignore during search (e.g., node_modules, .git)
  -h, --help             Help for filecon

Additional Features:
  - Support for .fileconignore file in the target directory to specify files and folders to ignore
  - Support for FILECON_EXTIGNORE environment variable to globally ignore file extensions

Examples:
  1. Run the interactive wizard:
     filecon

  2. Concatenate all .go files in the current directory into filecon_output.txt:
     filecon --dir=. --ext=go --out=output.txt

  3. Concatenate all .js files in /path/to/dir into filecon_result.js, removing extra spaces:
     filecon --dir=/path/to/dir --ext=.js --out=result.js --remove-spaces

  4. Concatenate all .py files in the current directory, ignoring .pyc files and the venv folder:
     filecon --dir=. --ext=py --ignore-ext=pyc --ignore-folders=venv,__pycache__

  5. Using .fileconignore:
     Create a file named .fileconignore in your target directory with patterns to ignore:
     # Comments are supported
     *.tmp
     *.bak
     node_modules/
     .git/

  6. Using environment variable:
     export FILECON_EXTIGNORE=.tmp,.bak,.test.js
     filecon --dir=. --ext=.js

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
	version        = "0.0.2"
	gitRepo        = "https://github.com/monzim/filecon"
	ignoreFileName = ".fileconignore"
	envIgnoreVar   = "FILECON_EXTIGNORE"
	developerURL   = "https://monzim.com"
)

var (
	dir           string
	fileType      string
	outputFile    string
	removeSpaces  bool
	ignoreExts    string
	ignoreFolders string
	ignoreInfo    []string // Used to collect information about ignore sources

	// Stats for reporting
	totalFilesScanned      int
	totalFilesConcatenated int
	totalOutputBytes       int64
	totalOutputLines       int
	filesIgnored           []string
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
			// Make sure all extensions have a dot
			for i, ext := range ignoreExtList {
				ext = strings.TrimSpace(ext)
				if ext != "" && !strings.HasPrefix(ext, ".") {
					ignoreExtList[i] = "." + ext
				}
			}

			ignoreFolderList := strings.Split(ignoreFolders, ",")

			// Check for environment variable
			ignoreInfo = []string{}
			if envIgnores := os.Getenv(envIgnoreVar); envIgnores != "" {
				envIgnoreList := strings.Split(envIgnores, ",")
				for _, ext := range envIgnoreList {
					ext = strings.TrimSpace(ext)
					if ext != "" {
						if !strings.HasPrefix(ext, ".") {
							ext = "." + ext
						}
						ignoreExtList = append(ignoreExtList, ext)
					}
				}
				ignoreInfo = append(ignoreInfo, fmt.Sprintf("%s environment variable", envIgnoreVar))
			}

			// Check for .fileconignore file
			ignoreFileExtList, ignoreFileFolderList, hasIgnoreFile := readIgnoreFile(dir)
			if hasIgnoreFile {
				ignoreExtList = append(ignoreExtList, ignoreFileExtList...)
				ignoreFolderList = append(ignoreFolderList, ignoreFileFolderList...)
				ignoreInfo = append(ignoreInfo, fmt.Sprintf("%s file", ignoreFileName))
			}

			if err := concatenateFiles(dir, fileType, outputFile, removeSpaces, ignoreExtList, ignoreFolderList); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			displaySummary(outputFile)
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
			t.Placeholder = "File extension (e.g., go, js, py - with or without dot)"
		case 2:
			t.Placeholder = "Output file (optional, default is filecon_<timestamp>.txt)"
		case 3:
			t.Placeholder = "Extensions to ignore (comma-separated, e.g., tmp,bak)"
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
				fileExtension := m.inputs[1].Value()
				// Make sure file extension starts with a dot
				if fileExtension != "" && !strings.HasPrefix(fileExtension, ".") {
					fileExtension = "." + fileExtension
				}

				ignoreExtList := strings.Split(m.inputs[3].Value(), ",")
				// Make sure all ignore extensions have a dot
				for i, ext := range ignoreExtList {
					ext = strings.TrimSpace(ext)
					if ext != "" && !strings.HasPrefix(ext, ".") {
						ignoreExtList[i] = "." + ext
					}
				}
				ignoreFolderList := strings.Split(m.inputs[4].Value(), ",")

				// Check for environment variable
				ignoreInfo = []string{}
				if envIgnores := os.Getenv(envIgnoreVar); envIgnores != "" {
					envIgnoreList := strings.Split(envIgnores, ",")
					for _, ext := range envIgnoreList {
						ext = strings.TrimSpace(ext)
						if ext != "" {
							ignoreExtList = append(ignoreExtList, ext)
						}
					}
					ignoreInfo = append(ignoreInfo, fmt.Sprintf("%s environment variable", envIgnoreVar))
				}

				// Check for .fileconignore file
				ignoreFileExtList, ignoreFileFolderList, hasIgnoreFile := readIgnoreFile(dir)
				if hasIgnoreFile {
					ignoreExtList = append(ignoreExtList, ignoreFileExtList...)
					ignoreFolderList = append(ignoreFolderList, ignoreFileFolderList...)
					ignoreInfo = append(ignoreInfo, fmt.Sprintf("%s file", ignoreFileName))
				}

				m.err = concatenateFiles(dir, fileExtension, outputFile, m.removeSpaces, ignoreExtList, ignoreFolderList)
				if m.err == nil {
					displaySummary(outputFile)
				}
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
		return "Press any key to exit."
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

// readIgnoreFile reads the .fileconignore file and returns lists of extensions and folders to ignore
func readIgnoreFile(dir string) ([]string, []string, bool) {
	ignoreFilePath := filepath.Join(dir, ignoreFileName)

	// Check if the file exists
	if _, err := os.Stat(ignoreFilePath); os.IsNotExist(err) {
		return nil, nil, false
	}

	// Read the file
	content, err := ioutil.ReadFile(ignoreFilePath)
	if err != nil {
		fmt.Printf("Warning: Found %s but couldn't read it: %v\n", ignoreFileName, err)
		return nil, nil, false
	}

	var extensions []string
	var folders []string

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Determine if it's a folder or extension
		if strings.HasPrefix(line, "/") || strings.HasSuffix(line, "/") || !strings.Contains(line, ".") {
			// It's a folder - clean up any slashes
			line = strings.Trim(line, "/")
			folders = append(folders, line)
		} else if strings.HasPrefix(line, "*.") {
			// It's a file extension pattern like "*.txt"
			extensions = append(extensions, strings.TrimPrefix(line, "*"))
		} else if strings.HasPrefix(line, ".") {
			// It's a file extension directly like ".txt"
			extensions = append(extensions, line)
		} else if strings.Contains(line, ".") && !strings.Contains(line, "/") && !strings.Contains(line, "\\") {
			// Likely a file extension without a dot prefix
			if !strings.HasPrefix(line, ".") {
				line = "." + line
			}
			extensions = append(extensions, line)
		} else {
			// Assume it's a specific filename or pattern
			// For now, we're treating all other patterns as extensions
			extensions = append(extensions, line)
		}
	}

	return extensions, folders, true
}

func concatenateFiles(dir, fileType, outputFile string, removeSpaces bool, ignoreExts, ignoreFolders []string) error {
	// Make sure fileType starts with a dot if it doesn't already
	if fileType != "" && !strings.HasPrefix(fileType, ".") {
		fileType = "." + fileType
	}

	// Reset stats
	totalFilesScanned = 0
	totalFilesConcatenated = 0
	totalOutputBytes = 0
	totalOutputLines = 0
	filesIgnored = []string{}

	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer outFile.Close()

	// Write the filecon signature at the top of the file
	signature := fmt.Sprintf("# Generated by File Concatenator (filecon) v%s\n# %s\n# Generated on: %s\n",
		version,
		gitRepo,
		time.Now().Format("2006-01-02 15:04:05"))

	// Add information about ignore sources if any
	if len(ignoreInfo) > 0 {
		signature += fmt.Sprintf("# Using ignore patterns from: %s\n", strings.Join(ignoreInfo, ", "))
		signature += "# Tool developed by: https://monzim.com"
	}

	signature += "\n"

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
					// Add to ignored files list
					filesIgnored = append(filesIgnored, fmt.Sprintf("Directory: %s", path))
					return filepath.SkipDir
				}
			}
			return nil
		}

		totalFilesScanned++

		// Skip files with extensions in the ignore list
		for _, ext := range ignoreExts {
			ext = strings.TrimSpace(ext)
			if ext != "" && strings.HasSuffix(info.Name(), ext) {
				filesIgnored = append(filesIgnored, path)
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

			header := fmt.Sprintf("# %s\n---\n", path)
			if _, err = outFile.WriteString(header); err != nil {
				return fmt.Errorf("error writing to output file: %v", err)
			}

			contentStr := string(content) + "\n\n"
			if _, err = outFile.WriteString(contentStr); err != nil {
				return fmt.Errorf("error writing file content to output file: %v", err)
			}

			totalFilesConcatenated++
			totalOutputBytes += int64(len(header) + len(contentStr))
			totalOutputLines += strings.Count(contentStr, "\n") + 2 // +2 for header lines
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

func displaySummary(outputFile string) {
	// Get output file info
	fileInfo, err := os.Stat(outputFile)
	fileSize := int64(0)
	if err == nil {
		fileSize = fileInfo.Size()
	}

	// Ensure the terminal respects newlines and spacing
	divider := strings.Repeat("─", 50)
	output := fmt.Sprintf(
		"✅ File concatenation completed | Total files scanned: %d | Files concatenated: %d | Output file: %s | Output size: %.2f KB | Total lines: %d\n",
		totalFilesScanned, totalFilesConcatenated, outputFile, float64(fileSize)/1024, totalOutputLines,
	)

	// Use raw `os.Stdout.WriteString()` to prevent Cobra formatting issues
	// output := "\n✅ File concatenation completed successfully!\n\n" +
	// 	divider + "\n" +
	// 	"\n📊 SUMMARY:\n\n" +
	// 	fmt.Sprintf("   Total files scanned:  %d\n", totalFilesScanned) +
	// 	fmt.Sprintf("   Files concatenated:   %d\n", totalFilesConcatenated) +
	// 	fmt.Sprintf("   Output file:          %s\n", outputFile) +
	// 	fmt.Sprintf("   Output size:          %.2f KB\n", float64(fileSize)/1024) +
	// 	fmt.Sprintf("   Total lines:          %d\n", totalOutputLines)

	// Append ignored sources if any
	if len(ignoreInfo) > 0 {
		output += "\n" + divider + "\n🔍 IGNORE SOURCES:\n"
		for _, source := range ignoreInfo {
			output += fmt.Sprintf("   - %s\n", source)
		}
	}

	// Append ignored files (limited to 5)
	if len(filesIgnored) > 0 {
		output += "\n" + divider + "\n⏭️ IGNORED FILES/DIRECTORIES:\n"
		showCount := len(filesIgnored)
		if showCount > 5 {
			showCount = 5
		}

		for i := 0; i < showCount; i++ {
			output += fmt.Sprintf("   - %s\n", filesIgnored[i])
		}

		if len(filesIgnored) > 5 {
			output += fmt.Sprintf("   ... and %d more\n", len(filesIgnored)-5)
		}
	}

	// Footer
	output += "\n" + divider + "\n\nThank you for using filecon!\n" +
		fmt.Sprintf("Developed by: %s\n\n", developerURL)

	// Force proper printing using os.Stdout.WriteString()
	os.Stdout.WriteString(output)
}
