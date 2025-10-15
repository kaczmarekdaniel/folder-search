// Package dirsearch provides directory searching and scanning functionality.
//
// This package implements recursive directory traversal with support for
// pattern matching, case-sensitive/insensitive search, and filtering of
// specific directory patterns.
package dirsearch

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

// DirSearch represents a directory search instance with configurable options.
// It provides methods to scan directories and find matches based on specified criteria.
type DirSearch struct {
	// Options contains the configuration for search operations
	Options *Options
}

// NewDirSearch creates a new DirSearch instance with default options.
//
// Default options include:
//   - No search pattern (matches all directories)
//   - Current directory (".") as start directory
//   - Case-insensitive search
//   - Ignoring "node_modules" directories
func NewDirSearch() *DirSearch {
	return &DirSearch{
		Options: DefaultOptions(),
	}
}

// ScanDirs scans the specified directory and returns all matching subdirectories.
//
// It updates the StartDir option and performs the search. Only direct child
// directories are returned (not nested subdirectories).
//
// Parameters:
//   - dir: the directory path to scan
//
// Returns a Result containing the list of matching directories or an error.
func (d *DirSearch) ScanDirs(dir string) Result {
	d.Options.StartDir = dir
	return Search(d.Options)
}

// Options configures the behavior of directory search operations.
type Options struct {
	// SearchPattern is the pattern to match against directory names.
	// Empty string matches all directories.
	SearchPattern string

	// StartDir is the directory where the search begins.
	StartDir string

	// CaseSensitive determines whether pattern matching is case-sensitive.
	CaseSensitive bool

	// IgnorePatterns is a list of directory names to skip during traversal.
	IgnorePatterns []string
}

// Result contains the outcome of a directory search operation.
type Result struct {
	// Directories is the list of matching directory paths (relative to StartDir)
	Directories []string

	// Error contains any error that occurred during the search
	Error error
}

// DefaultOptions returns the default search options.
//
// Returns Options configured with:
//   - Empty search pattern (matches all)
//   - Current directory as start directory
//   - Case-insensitive matching
//   - node_modules in ignore list
func DefaultOptions() *Options {
	return &Options{
		SearchPattern:  "",
		StartDir:       ".",
		CaseSensitive:  false,
		IgnorePatterns: []string{"node_modules"},
	}
}

// Search performs a directory search with the given options.
//
// It reads only the immediate child directories of opts.StartDir,
// applying the following rules:
//   - Skips .git directories automatically
//   - Skips directories matching patterns in opts.IgnorePatterns
//   - Matches directory names against opts.SearchPattern (if provided)
//   - Returns only direct child directories (not nested subdirectories)
//   - Returns relative paths from opts.StartDir
//
// The function uses os.ReadDir for non-recursive, efficient directory reading.
// Permission errors and other read errors are silently skipped.
//
// Parameters:
//   - opts: configuration options for the search
//
// Returns a Result with matching directories or an error.
func Search(opts *Options) Result {
	foundDirs := []string{}

	// Prepare pattern for search
	var pattern string
	if opts.CaseSensitive {
		pattern = opts.SearchPattern
	} else {
		pattern = strings.ToLower(opts.SearchPattern)
	}

	nameProvided := opts.SearchPattern != ""

	// Read only immediate children (non-recursive)
	entries, err := os.ReadDir(opts.StartDir)
	if err != nil {
		return Result{
			Directories: foundDirs,
			Error:       err,
		}
	}

	// Process each entry
	for _, entry := range entries {
		// Skip non-directories
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Skip .git directories
		if strings.HasPrefix(name, ".git") {
			continue
		}

		// Skip directories in ignore patterns
		if slices.Contains(opts.IgnorePatterns, name) {
			continue
		}

		// Check if it matches the search pattern
		var matches bool
		if !nameProvided {
			matches = true
		} else if opts.CaseSensitive {
			matches = strings.Contains(name, pattern)
		} else {
			matches = strings.Contains(strings.ToLower(name), pattern)
		}

		if matches {
			foundDirs = append(foundDirs, name)
		}
	}

	return Result{
		Directories: foundDirs,
		Error:       nil,
	}
}

// PrintResults prints the search results in a formatted, human-readable way.
//
// It outputs:
//   - An error message if the result contains an error
//   - The count of directories found
//   - A numbered list of all directory paths
//
// This function is primarily useful for CLI debugging and testing.
// The UI package should not use this function.
func PrintResults(result Result) {
	if result.Error != nil {
		fmt.Printf("Error walking directory: %v\n", result.Error)
		return
	}

	fmt.Printf("Found %d directories:\n", len(result.Directories))
	for i, dir := range result.Directories {
		fmt.Printf("%d. %s\n", i+1, dir)
	}
}
