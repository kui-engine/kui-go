//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_platform_04_file_dialog.go; edit that file, not this copy.
// Host logic of examples/native/platform/04_file_dialog/main.cpp: Open File / Save File / Select
// Folder show the native dialogs and write the chosen path, or "... cancelled.", into #result.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	show := func(path string, ok bool, err error, cancelled string) {
		if err != nil || !ok {
			path = cancelled
		}
		if result, err := doc.GetElementByID("result"); err == nil && result != nil {
			_ = result.SetTextContent(path)
			result.Close()
		}
	}
	handlers := []struct {
		id string
		fn func()
	}{
		{"openBtn", func() {
			path, ok, err := kui.OpenFileDialog(kui.FileDialogOptions{
				Title: "Open a document", Filters: []string{"*.html", "*.css", "*.txt"}})
			show(path, ok, err, "Open cancelled.")
		}},
		{"saveBtn", func() {
			path, ok, err := kui.SaveFileDialog(kui.FileDialogOptions{
				Title: "Save a document", DefaultPath: "example.html", Filters: []string{"*.html"}})
			show(path, ok, err, "Save cancelled.")
		}},
		{"folderBtn", func() {
			path, ok, err := kui.SelectFolderDialog("Select a folder")
			show(path, ok, err, "Folder selection cancelled.")
		}},
	}
	for _, handler := range handlers {
		button, err := doc.GetElementByID(handler.id)
		if err != nil || button == nil {
			return fmt.Errorf("没有 #%s: %v", handler.id, err)
		}
		ctx.Keep(button)
		fn := handler.fn
		if _, err := button.AddEventListener("click", func(*kui.Event) { fn() }); err != nil {
			return err
		}
	}
	return nil
}
