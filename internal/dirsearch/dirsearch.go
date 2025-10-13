// Package dirsearch provides directory searching and scanning functionality.
//
// This package implements recursive directory traversal with support for
// pattern matching, case-sensitive/insensitive search, and filtering of
// specific directory patterns.
package dirsearch

import (
	"fmt"
	"os"
	"path/filepath"
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
// It recursively traverses the directory tree starting from opts.StartDir,
// applying the following rules:
//   - Skips .git directories automatically
//   - Skips directories matching patterns in opts.IgnorePatterns
//   - Matches directory names against opts.SearchPattern (if provided)
//   - Returns only direct child directories (not nested subdirectories)
//   - Returns relative paths from opts.StartDir
//
// The function uses filepath.WalkDir for efficient directory traversal.
// Errors during traversal are skipped with filepath.SkipDir.
//
// Parameters:
//   - opts: configuration options for the search
//
// Returns a Result with matching directories or an error.
func Search(opts *Options) Result {
	var foundDirs []string
	var searchErr error
	// Prepare pattern for search
	var pattern string
	if opts.CaseSensitive {
		pattern = opts.SearchPattern
	} else {
		pattern = strings.ToLower(opts.SearchPattern)
	}

	nameProvided := opts.SearchPattern != ""

	// Walk through all directories
	searchErr = filepath.WalkDir(opts.StartDir, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return filepath.SkipDir
		}

		if info.IsDir() {
			matched := strings.HasPrefix(info.Name(), ".git")

			if matched {
				return filepath.SkipDir
			}

			if slices.Contains(opts.IgnorePatterns, info.Name()) {
				return filepath.SkipDir
			}
		}

		// If it's a directory and matches our pattern (if provided)
		if info.IsDir() {
			var matches bool
			if !nameProvided {
				matches = true
			} else if opts.CaseSensitive {
				matches = strings.Contains(info.Name(), pattern)
			} else {
				matches = strings.Contains(strings.ToLower(info.Name()), pattern)
			}

			if matches {
				relativePath, err := filepath.Rel(opts.StartDir, path)
				if err != nil {
					relativePath = path
				}
				if relativePath == "." {
					return nil
				}

				if strings.Contains(relativePath, "/") {
					return nil
				}

				// Add to our slice
				foundDirs = append(foundDirs, relativePath)
			}
		}

		return nil
	})

	return Result{
		Directories: foundDirs,
		Error:       searchErr,
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
