package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/mikkel-andersen/nota/adapter/editor"
	"github.com/mikkel-andersen/nota/adapter/repository"
	"github.com/mikkel-andersen/nota/adapter/scanner"
	"github.com/mikkel-andersen/nota/cmd"
	"github.com/mikkel-andersen/nota/usecase"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	safe := strings.NewReplacer("/", "_", "\\", "_", ":", "_").Replace(cwd)
	repoDir := filepath.Join(xdg.DataHome, "nota", safe)
	baseDir := filepath.Join(xdg.DataHome, "nota")

	repo, err := repository.NewJSONRepo(repoDir)
	if err != nil {
		panic(err)
	}

	svc := usecase.NewNoteService(
		repo,
		scanner.NewJSONScanner(baseDir),
		editor.NewOSEditor(),
	)

	cmd.Execute(svc)
}
