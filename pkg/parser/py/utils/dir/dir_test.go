package dir

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindPyModules(t *testing.T) {
	// Create a temporary directory for testing
	rootDir := createTempDirectory(t)
	defer os.RemoveAll(rootDir)

	// Create some directories and __init__.py files
	createDirectoryWithInitPy(rootDir, "package1")
	createDirectoryWithInitPy(rootDir, "package2")
	createDirectoryWithInitPy(rootDir, "package3")
	createEmptyDirectory(rootDir, "emptyDir")

	// Call the FindPyModules function
	packageNames, err := FindRootModules(rootDir)
	if err != nil {
		t.Fatalf("Error while finding Python modules: %v", err)
	}

	// Check the expected package names
	expectedPackageNames := []string{"package1", "package2", "package3"}
	for _, expectedName := range expectedPackageNames {
		found := false
		for _, packageName := range packageNames {
			if packageName == expectedName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected package name %s not found in result", expectedName)
		}
	}

	// Check that the empty directory is not included
	for _, packageName := range packageNames {
		if packageName == "emptyDir" {
			t.Errorf("Empty directory 'emptyDir' should not be included in result")
		}
	}
}

func createTempDirectory(t *testing.T) string {
	tempDir := t.TempDir()
	return tempDir
}

func createDirectoryWithInitPy(rootDir, packageName string) {
	packageDir := filepath.Join(rootDir, packageName)
	if err := os.Mkdir(packageDir, os.ModePerm); err != nil {
		panic(err)
	}

	initPyFile := filepath.Join(packageDir, "__init__.py")
	if _, err := os.Create(initPyFile); err != nil {
		panic(err)
	}
}

func createEmptyDirectory(rootDir, directoryName string) {
	directoryPath := filepath.Join(rootDir, directoryName)
	if err := os.Mkdir(directoryPath, os.ModePerm); err != nil {
		panic(err)
	}
}
