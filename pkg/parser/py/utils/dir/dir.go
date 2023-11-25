package dir

import (
	"os"
	"path/filepath"
	"strings"
)

func FindTopLevelModules(rootDir string) (map[string]string, error) {
	packageNames := make(map[string]string, 0)
	processedDirs := make(map[string]bool)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && !strings.Contains(path, "__init__.py") {
			// Check if the directory contains an __init__.py file
			initPyFile := filepath.Join(path, "__init__.py")
			_, err := os.Stat(initPyFile)

			if err != nil {
				return nil
			}
			// Found an __init__.py file
			packageName := filepath.Base(path)
			// Check if the directory is a top-level directory
			parentDir := filepath.Dir(path)
			if _, exists := processedDirs[parentDir]; !exists {
				packageNames[packageName] = path
			}
			processedDirs[path] = true
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return packageNames, nil
}
