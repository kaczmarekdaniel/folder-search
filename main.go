package main

import (
	"os"

	"github.com/kaczmarekdaniel/folder-search/internal/app"
	"github.com/kaczmarekdaniel/folder-search/internal/ui"
)

func main() {
	app, err := app.NewApplication()
	if err != nil {
		os.Exit(1)

	}

	ui.InitUI(app)

}

//  Create app struct, but what should it do?
//   - hold references to search (rename it) and ui
//   - make sure it's initialised only once
//
//  How flow should look like?
//   -> main func is called -> read flags from command -> pass search in folder func & options from flags to ui ->
//   ui then calls search on folder X when needed
