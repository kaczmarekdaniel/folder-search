package ui

import (
	"fmt"
	"io"
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

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

// Types
type item string

type model struct {
	requestChan chan string
	resultChan  chan dirsearch.Result
	list        list.Model
	choice      string
	quitting    bool
	responses   int
	search      func(dir string) dirsearch.Result
	prevDir     string
	currentDir  string
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

func scanInBackground(requestChan chan string, resultChan chan dirsearch.Result, searchFunc func(dir string) dirsearch.Result) {
	for dir := range requestChan {
		result := searchFunc(dir)
		resultChan <- result
	}
}

func waitForResults(resultChan chan dirsearch.Result) tea.Cmd {
	return func() tea.Msg {
		result := <-resultChan
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
			m.quitting = true
			return m, tea.Quit
		case "right":

			i, _ := m.list.SelectedItem().(item)
			m.currentDir = m.currentDir + "/" + string(i)
			// Send request to scan the new directory
			m.requestChan <- m.currentDir
		case "left":
			parentDir := filepath.Dir(m.currentDir)
			m.currentDir = parentDir
			// Send request to scan the parent directory
			m.requestChan <- m.currentDir

		case "s":
			parentDir := filepath.Dir(m.currentDir)
			m.currentDir = parentDir
			// Send request to scan the parent directory
			m.requestChan <- m.currentDir

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			return m, tea.Quit
		}
	case responseMsg:
		result := msg.result
		if result.Error == nil {
			m.list.SetItems(stringsToItems(result.Directories))
			height := int(math.Min(float64(len(result.Directories)+8), 24))
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

	save := key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "save path"),
	)
	favourites := key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "search"),
	)

	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{save, favourites}
	}

	listView := m.list.View()
	footer := "\n Press q to exit\n"

	return listView + footer
}

func InitUI(app *app.Application) {
	result := app.Dirsearch.ScanDirs(".")
	const title = ""
	if result.Error != nil {
		fmt.Printf("Error: %v\n", result.Error)
		os.Exit(1)
	}

	const defaultWidth = 20
	items := stringsToItems(result.Directories)
	height := int(math.Min(float64(len(items)+8), 64))
	l := list.New(items, itemDelegate{}, defaultWidth, height)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	// l.SetFilterText("")

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return
	}

	requestChan := make(chan string)
	resultChan := make(chan dirsearch.Result)

	go scanInBackground(requestChan, resultChan, app.Dirsearch.ScanDirs)

	m := model{
		list:        l,
		currentDir:  currentDir,
		requestChan: requestChan,
		resultChan:  resultChan,
		search:      app.Dirsearch.ScanDirs,
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
