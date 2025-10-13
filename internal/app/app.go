// Package app provides the core application structure and initialization.
//
// This package coordinates between different components of the folder-search application,
// managing the directory search functionality and application-wide logging.
package app

import (
	"log/slog"
	"os"

	"github.com/kaczmarekdaniel/folder-search/internal/dirsearch"
)

// Application represents the core application structure that holds
// references to all major components including directory search and logging.
type Application struct {
	// Dirsearch handles directory scanning and searching operations
	Dirsearch *dirsearch.DirSearch

	// Logger provides structured logging throughout the application
	Logger *slog.Logger
}

// NewApplication creates and initializes a new Application instance with default configuration.
//
// It sets up:
//   - A structured logger using slog with INFO level output to stderr
//   - A directory search instance with default options
//
// Returns an error if initialization fails (currently always returns nil error).
func NewApplication() (*Application, error) {
	// Create structured logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	searchDir := dirsearch.NewDirSearch()

	app := &Application{
		Dirsearch: searchDir,
		Logger:    logger,
	}

	logger.Info("application initialized")
	return app, nil
}
