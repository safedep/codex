package dir

import (
	"os"
	"path/filepath"
	"strings"
)

func FindRootModules(rootDir string) ([]string, error) {
	packageNames := make([]string, 0)
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && !strings.Contains(path, "__init__.py") {
			// Check if the directory contains an __init__.py file
			initPyFile := filepath.Join(path, "__init__.py")
			_, err := os.Stat(initPyFile)

			if err == nil {
				// Found an __init__.py file
				packageName := filepath.Base(path)
				packageNames = append(packageNames, packageName)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return packageNames, nil
}
