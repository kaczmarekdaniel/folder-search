# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`folder-search` is a terminal-based interactive directory navigator built with Go and Charm's Bubble Tea TUI framework. It provides real-time directory browsing with keyboard navigation.

## Commands

### Build and Run
```bash
# Run the application
go run main.go

# Build the binary
go build -o folder-search

# Install dependencies
go mod download

# Tidy dependencies
go mod tidy
```

### Testing
Currently no tests are implemented in the codebase.

## Architecture

### Core Components

The application follows a three-layer architecture:

1. **main.go**: Entry point that initializes the application and starts the UI
2. **internal/app**: Application layer that coordinates between components
3. **internal/dirsearch**: Business logic for directory scanning and filtering
4. **internal/ui**: TUI implementation using Bubble Tea framework

### Key Architectural Patterns

**Application Initialization Flow**:
- `main.go` creates an `app.Application` instance
- `app.Application` holds a reference to `dirsearch.DirSearch`
- UI is initialized by passing the application instance to `ui.InitUI()`

**Asynchronous Directory Scanning**:
The UI uses a channel-based architecture for non-blocking directory scans:
- `requestChan`: Sends directory paths to scan
- `resultChan`: Receives `dirsearch.Result` with found directories
- `scanInBackground()` goroutine processes scan requests continuously
- `waitForResults()` returns a Bubble Tea command that waits for results

**UI Event Loop** (internal/ui/ui.go:109):
The `Update()` method handles:
- Keyboard events: navigation (left/right arrows), selection (enter), quit (q/ctrl+c)
- Directory navigation: right arrow enters folder, left arrow goes to parent
- Response messages: updates list when new scan results arrive

### Directory Search Logic

**Search Implementation** (internal/dirsearch/dirsearch.go:78):
- Uses `filepath.WalkDir` to traverse directories
- Filters out `.git` directories and patterns in `IgnorePatterns` (defaults to `node_modules`)
- Only returns direct child directories (not nested subdirectories) - see line 131-133
- Supports case-sensitive and case-insensitive search
- Returns relative paths from the starting directory

**Search Options** (internal/dirsearch/dirsearch.go:28):
- `SearchPattern`: Pattern to match directory names (empty matches all)
- `StartDir`: Root directory to start search
- `CaseSensitive`: Boolean for case sensitivity
- `IgnorePatterns`: Slice of directory names to skip

### UI State Management

The `model` struct (internal/ui/ui.go:31) maintains:
- `currentDir`: Current directory being displayed
- `list`: Bubble Tea list component with found directories
- `requestChan`/`resultChan`: Communication channels for async scanning
- `search`: Function reference to `app.Dirsearch.ScanDirs`
- `showSaved`: Toggle for saved paths view (not fully implemented)

### Known Incomplete Features

Based on code inspection:
- Saved paths functionality (s/f keys) is partially implemented but not persistent
- Comments in main.go:21-27 indicate ongoing architectural considerations
