package exampleutil

import (
	"os"
	"path/filepath"

	kui "github.com/kui-engine/kui-go/v3"
)

// Argument returns an explicit command-line override or a package-local
// default. Prebuilt SDK examples can therefore run directly from bin/ while
// source consumers can still point at another runtime or asset.
func Argument(index int, fallback string) string {
	if len(os.Args) > index && os.Args[index] != "" {
		return os.Args[index]
	}
	return fallback
}

func ExecutableSibling(name string) string {
	executable, err := os.Executable()
	if err != nil {
		return name
	}
	return filepath.Join(filepath.Dir(executable), name)
}

func RuntimePath() string {
	return Argument(1, ExecutableSibling(kui.RuntimeLibraryName))
}
