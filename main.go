package main

import (
	"fmt"
	"os"

	"github.com/kaczmarekdaniel/folder-search/internal/app"
	"github.com/kaczmarekdaniel/folder-search/internal/ui"
)

func main() {
	app, err := app.NewApplication()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing application: %v\n", err)
		os.Exit(1)
	}

	app.Logger.Info("starting UI")
	if err := ui.InitUI(app); err != nil {
		app.Logger.Error("failed to run UI", "error", err)
		fmt.Fprintf(os.Stderr, "Error running UI: %v\n", err)
		os.Exit(1)
	}
	app.Logger.Info("application exiting normally")
}
