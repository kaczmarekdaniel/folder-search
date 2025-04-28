package ui

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	sub       chan int
	list      list.Model
	choice    string
	quitting  bool
	responses int
}

type responseMsg struct {
	RandomNumber int
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

func waitForActivity(sub chan int) tea.Cmd {
	return func() tea.Msg {
		num := <-sub
		return responseMsg{RandomNumber: num}
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		waitForActivity(m.sub),
	)
}

// Update handles different types of events around the list and returns an updated model and command.
//
// It processes window size changes, keyboard events, and response messages using nested
// switch statements. Specific key actions include:
//   - q/ctrl+c: quit the application
//   - right: send a random number to the subscription channel
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
			randomNum := rand.Intn(1000)
			m.sub <- randomNum
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			return m, tea.Quit
		}
	case responseMsg:

		i, _ := m.list.SelectedItem().(item)

		newFolderItems := []string{string(i)}
		m.list.SetItems(stringsToItems(newFolderItems))

		height := int(math.Min(float64(len(newFolderItems)+6), 24))
		m.list.SetHeight(height)

		return m, waitForActivity(m.sub)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.choice != "" {
		return quitTextStyle.Render(fmt.Sprintf("%s? Sounds good to me.", m.choice))
	}
	if m.quitting {
		return quitTextStyle.Render("Not hungry? That's cool.")
	}

	header := fmt.Sprintf("\n Events received: %d\n", m.responses)
	listView := m.list.View()
	footer := "\n Press q to exit\n"

	return header + listView + footer
}

func Print(stringList []string, title string) {
	const defaultWidth = 20
	items := stringsToItems(stringList)
	height := int(math.Min(float64(len(items)+6), 64))
	l := list.New(items, itemDelegate{}, defaultWidth, height)
	if title == "" {
		l.Title = "Searching in a filetree"
	} else {
		l.Title = title
	}
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := model{list: l, sub: make(chan int)}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
