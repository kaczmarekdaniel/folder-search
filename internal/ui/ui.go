// Package ui implements the terminal user interface using the Bubble Tea framework.
//
// This package provides an interactive directory browser that allows users to:
//   - Navigate through directory hierarchies
//   - View directory contents in real-time
//   - Select directories with keyboard navigation
//   - Handle errors gracefully with user-friendly messages
//
// The UI runs asynchronously, scanning directories in the background without
// blocking user interaction.
package ui

import (
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kaczmarekdaniel/folder-search/internal/app"
	"github.com/kaczmarekdaniel/folder-search/internal/dirsearch"
)

const (
	// UI dimension constants
	defaultListWidth      = 20
	listHeightPadding     = 8
	maxListHeight         = 64
	maxDynamicListHeight  = 24

	// Style constants
	titleMarginLeft       = 2
	itemPaddingLeft       = 4
	selectedItemPadding   = 2
	quitTextTopMargin     = 1
	quitTextBottomMargin  = 2
	quitTextLeftMargin    = 4
	helpBottomPadding     = 1
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(titleMarginLeft)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(itemPaddingLeft)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(selectedItemPadding).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(itemPaddingLeft)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(itemPaddingLeft).PaddingBottom(helpBottomPadding)
	quitTextStyle     = lipgloss.NewStyle().Margin(quitTextTopMargin, 0, quitTextBottomMargin, quitTextLeftMargin)
)

// Types
type item string

type model struct {
	requestChan chan string
	resultChan  chan dirsearch.Result
	doneChan    chan struct{}
	list        list.Model
	choice      string
	quitting    bool
	responses   int
	search      func(dir string) dirsearch.Result
	prevDir     string
	currentDir  string
	err         error
	logger      *slog.Logger
}

type responseMsg struct {
	result dirsearch.Result
}

type itemDelegate struct{}

// Helpers
func (i item) FilterValue() string { return "" }

func stringsToItems(strs []string) []list.Item {
	items := make([]list.Item, 0, len(strs))
	for _, s := range strs {
		items = append(items, item(s))
	}
	return items
}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)
	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}
	fmt.Fprint(w, fn(str))
}

func scanInBackground(requestChan chan string, resultChan chan dirsearch.Result, doneChan chan struct{}, searchFunc func(dir string) dirsearch.Result) {
	for {
		select {
		case <-doneChan:
			close(requestChan)
			close(resultChan)
			return
		case dir := <-requestChan:
			result := searchFunc(dir)
			select {
			case resultChan <- result:
			case <-doneChan:
				close(requestChan)
				close(resultChan)
				return
			}
		}
	}
}

func waitForResults(resultChan chan dirsearch.Result) tea.Cmd {
	return func() tea.Msg {
		result, ok := <-resultChan
		if !ok {
			// Channel closed, return empty result
			return responseMsg{result: dirsearch.Result{}}
		}
		return responseMsg{result: result}
	}
}

func (m model) Init() tea.Cmd {
	m.requestChan <- m.currentDir
	return waitForResults(m.resultChan)
}

// Update handles different types of events around the list and returns an updated model and command.
//
// It processes window size changes, keyboard events, and response messages using nested
// switch statements. Specific key actions include:
//   - q/ctrl+c: quit the application
//   - right: enter the higlighted folder
//   - left: go to parent folder
//   - enter: select the current item and quit
//
// Response messages trigger addition of new items to the list.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.logger.Info("user quit application")
			m.quitting = true
			close(m.doneChan)
			return m, tea.Quit
		case "right":
			i, _ := m.list.SelectedItem().(item)
			m.currentDir = m.currentDir + "/" + string(i)
			m.logger.Debug("navigating into directory", "dir", m.currentDir)
			// Send request to scan the new directory
			m.requestChan <- m.currentDir
		case "left":
			parentDir := filepath.Dir(m.currentDir)
			m.currentDir = parentDir
			m.logger.Debug("navigating to parent directory", "dir", m.currentDir)
			// Send request to scan the parent directory
			m.requestChan <- m.currentDir
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			close(m.doneChan)
			return m, tea.Quit
		}
	case responseMsg:
		result := msg.result
		if result.Error != nil {
			m.logger.Error("directory scan failed", "error", result.Error, "dir", m.currentDir)
			m.err = result.Error
		} else {
			m.logger.Debug("directory scan completed", "dir", m.currentDir, "count", len(result.Directories))
			m.err = nil
			m.list.SetItems(stringsToItems(result.Directories))
			height := int(math.Min(float64(len(result.Directories)+listHeightPadding), maxDynamicListHeight))
			m.list.SetHeight(height)
		}
		return m, waitForResults(m.resultChan)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	m.list.Title = m.currentDir

	if m.choice != "" {
		return quitTextStyle.Render(fmt.Sprintf("%s? navigating to %s", m.choice, m.choice))
	}
	if m.quitting {
		return quitTextStyle.Render("See ya later, aligator")
	}

	// Display error if present
	if m.err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Margin(1, 2)
		return errorStyle.Render(fmt.Sprintf("Error: %v\nPress q to quit", m.err))
	}

	enter := key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open"),
	)

	left := key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "parent dir"),
	)

	right := key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "enter dir"),
	)

	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{left, right, enter}
	}

	return m.list.View()
}

// InitUI initializes and runs the terminal user interface.
//
// This function:
//   1. Performs an initial directory scan of the current directory
//   2. Sets up the Bubble Tea list component with the results
//   3. Creates background goroutines for async directory scanning
//   4. Starts the Bubble Tea event loop
//   5. Blocks until the user quits the application
//
// The UI provides keyboard controls for navigation:
//   - Up/Down or j/k: Navigate through directories
//   - Right or l: Enter selected directory
//   - Left or h: Go to parent directory
//   - Enter: Select directory and exit
//   - q or Ctrl+C: Quit application
//
// Parameters:
//   - app: The application instance containing the directory searcher and logger
//
// Returns an error if:
//   - Initial directory scan fails
//   - Current working directory cannot be determined
//   - Bubble Tea program encounters an error
func InitUI(app *app.Application) error {
	app.Logger.Info("initializing UI")
	result := app.Dirsearch.ScanDirs(".")
	const title = ""
	if result.Error != nil {
		app.Logger.Error("initial directory scan failed", "error", result.Error)
		return fmt.Errorf("initial directory scan failed: %w", result.Error)
	}
	app.Logger.Debug("initial scan completed", "count", len(result.Directories))

	items := stringsToItems(result.Directories)
	height := int(math.Min(float64(len(items)+listHeightPadding), maxListHeight))
	l := list.New(items, itemDelegate{}, defaultListWidth, height)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	// l.SetFilterText("")

	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	requestChan := make(chan string)
	resultChan := make(chan dirsearch.Result)
	doneChan := make(chan struct{})

	go scanInBackground(requestChan, resultChan, doneChan, app.Dirsearch.ScanDirs)

	m := model{
		list:        l,
		currentDir:  currentDir,
		requestChan: requestChan,
		resultChan:  resultChan,
		doneChan:    doneChan,
		search:      app.Dirsearch.ScanDirs,
		logger:      app.Logger,
	}

	app.Logger.Info("starting UI event loop")

	if _, err := tea.NewProgram(m).Run(); err != nil {
		return fmt.Errorf("failed to run UI program: %w", err)
	}

	return nil
}
