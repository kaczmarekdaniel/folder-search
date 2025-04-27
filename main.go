package main

import (
	"fmt"
	"os"

	"github.com/kaczmarekdaniel/folder-search/internal/dirsearch"
	"github.com/kaczmarekdaniel/folder-search/internal/ui"
)

func main() {
	opts := dirsearch.NewOptionsFromFlags()

	fmt.Printf("Searching directories in: %s\n", opts.StartDir)
	if opts.SearchPattern == "" {
		fmt.Println("No name pattern provided. Showing all directories.")
	}
	fmt.Printf("Ignoring directories: %v\n", opts.IgnorePatterns)

	result := dirsearch.Search(opts)

	if result.Error != nil {
		fmt.Printf("Error: %v\n", result.Error)
		os.Exit(1)
	}

	ui.Print(result.Directories, opts.SearchPattern)
}
