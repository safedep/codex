package dir

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/safedep/dry/log"
)

func FindTopLevelModules(rootDir string) (map[string]string, error) {
	packageNames := make(map[string]string, 0)
	processedDirs := make(map[string]bool)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		isTopLevelTxt := (strings.HasSuffix(path, "egg-info/top_level.txt") ||
			strings.HasSuffix(path, "dist-info/top_level.txt"))

		if !info.IsDir() && isTopLevelTxt {
			pkgs, err := ReadAllLines(path)
			if err != nil {
				log.Debugf("Error while reading top_level.txt file.. %s", err)
				return nil
			}
			for _, pkg := range pkgs {
				relativePath, _ := RelativePath(rootDir, path)
				packageNames[pkg] = relativePath
			}
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
				relativePath, _ := RelativePath(rootDir, path)
				packageNames[packageName] = relativePath
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

func ReadAllLines(filepath string) ([]string, error) {
	var lines []string
	readFile, err := os.Open(filepath)

	if err != nil {
		return lines, err
	}
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		lines = append(lines, fileScanner.Text())
	}

	return lines, nil
}

func RelativePath(basePath, fullPath string) (string, error) {
	// Clean and normalize the paths to ensure consistency
	basePath = filepath.Clean(basePath)
	fullPath = filepath.Clean(fullPath)

	// Check if the full path is inside the base path
	if !strings.HasPrefix(fullPath, basePath) {
		return "", fmt.Errorf("full path is not inside the base path")
	}

	// Calculate the relative path
	relativePath := strings.TrimPrefix(fullPath, basePath)
	relativePath = strings.TrimPrefix(relativePath, string(filepath.Separator))

	return relativePath, nil
}

func SplitAndGetLeftMost(path, sep string) string {
	// Split the path using the separator
	parts := strings.Split(path, sep)

	// Return the leftmost item (first element)
	if len(parts) > 0 {
		return parts[0]
	}

	// If there are no parts, return an empty string
	return ""
}
