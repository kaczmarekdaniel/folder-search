// dirsearch/dirsearch.go
package dirsearch

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type DirSearch struct {
	Options *DirSearchOptions
}

func NewDirSearchType() *DirSearch {
	return &DirSearch{
		Options: DefaultOptions(),
	}
}

func (d *DirSearch) ScanDirs(dir string) Result {
	d.Options.StartDir = dir
	return Search(d.Options)
}

type DirSearchOptions struct {
	SearchPattern  string
	StartDir       string
	CaseSensitive  bool
	IgnorePatterns []string
}

type TSearch interface {
	Search() Result
}

// Result contains search results
type Result struct {
	Directories []string
	Error       error
}

// DefaultOptions returns the default search options
func DefaultOptions() *DirSearchOptions {
	return &DirSearchOptions{
		SearchPattern:  "",
		StartDir:       ".",
		CaseSensitive:  false,
		IgnorePatterns: []string{"node_modules"},
	}
}

func NewOptionsFromFlags() *DirSearchOptions {
	opts := DefaultOptions()

	searchPattern := flag.String("name", "", "Search pattern for directory names")
	startDir := flag.String("dir", ".", "Starting directory for search")
	caseSensitive := flag.Bool("case-sensitive", false, "Enable case-sensitive search")

	flag.Parse()

	flag.Visit(func(f *flag.Flag) {
		if f.Name == "name" {
			opts.SearchPattern = *searchPattern
		}
	})

	opts.StartDir = *startDir
	opts.CaseSensitive = *caseSensitive

	return opts
}

// Search performs directory search with the given options
// func (wh *WorkoutHandler) HandleGetWorkoutByID(w http.ResponseWriter, r *http.Request) {
func Search(opts *DirSearchOptions) Result {
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

// PrintResults prints the search results in a formatted way
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

