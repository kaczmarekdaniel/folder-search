# folder-search

An interactive terminal-based directory navigator built with Go and Bubble Tea.

## Description

folder-search is a TUI (Terminal User Interface) application that allows you to browse and navigate through directories in real-time. It provides a fast, keyboard-driven interface for exploring folder structures.

## Features

- Real-time directory browsing with keyboard navigation
- Asynchronous directory scanning for responsive UI
- Navigate into subdirectories and back to parent directories
- Filter directories by name (case-sensitive and case-insensitive options)
- Automatic filtering of `.git` and `node_modules` directories
- Clean, minimal interface using Charm's Bubble Tea framework

## Requirements

- Go 1.24.1 or higher

## Installation

```bash
# Clone the repository
git clone https://github.com/kaczmarekdaniel/folder-search.git
cd folder-search

# Install dependencies
go mod download

# Build the application
go build -o folder-search

# Run the application
./folder-search
```

Or run directly without building:

```bash
go run main.go
```

## Usage

Launch the application from any directory to start browsing:

```bash
./folder-search
```

### Keyboard Controls

- **Up/Down arrows** or **j/k**: Navigate through the list of directories
- **Right arrow**: Enter the selected directory
- **Left arrow**: Go to parent directory
- **Enter**: Select the current directory and exit
- **s**: Save current path (feature in development)
- **f**: Toggle saved paths view (feature in development)
- **q** or **Ctrl+C**: Quit the application

## How It Works

The application starts in the current working directory and displays all immediate subdirectories. As you navigate with the arrow keys, it dynamically scans directories in the background using Go channels and goroutines, ensuring the interface remains responsive even when scanning large directory structures.

The search algorithm:
- Only shows direct child directories (not nested subdirectories)
- Automatically filters out `.git` directories
- Skips `node_modules` by default
- Returns relative paths from the starting directory

## Configuration

Default search options can be modified in `internal/dirsearch/dirsearch.go`:

```go
func DefaultOptions() *DirSearchOptions {
    return &DirSearchOptions{
        SearchPattern:  "",
        StartDir:       ".",
        CaseSensitive:  false,
        IgnorePatterns: []string{"node_modules"},
    }
}
```

## Project Structure

```
folder-search/
├── main.go                          # Application entry point
├── internal/
│   ├── app/
│   │   └── app.go                   # Application coordinator
│   ├── dirsearch/
│   │   └── dirsearch.go            # Directory search logic
│   └── ui/
│       └── ui.go                    # Terminal UI implementation
├── go.mod
└── go.sum
```

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions for terminal UIs

## License

This project is available under the terms specified by the repository owner.
