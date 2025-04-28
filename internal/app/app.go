package app

import "github.com/kaczmarekdaniel/folder-search/internal/dirsearch"

type Application struct {
	Dirsearch *dirsearch.DirSearch
}

func NewApplication() (*Application, error) {

	search_dir := dirsearch.NewDirSearchType()

	app := &Application{
		Dirsearch: search_dir,
	}

	return app, nil
}
