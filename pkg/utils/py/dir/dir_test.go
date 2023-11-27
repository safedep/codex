package dir

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindPyModules(t *testing.T) {
	// Create a temporary directory for testing
	rootDir := createTempDirectory(t)
	defer os.RemoveAll(rootDir)

	// Create some directories and __init__.py files
	/*
		rootDir/emptyDir1/
		rootDir/emptyDir1/package1/__init__.py
		rootDir/emptyDir1/package1/shouldnotpackage1/__init__.py
		rootDir/emptyDir1/emptyDir2
		rootDir/emptyDir1/emptyDir2/emptyDir3
		rootDir/emptyDir1/emptyDir2/emptyDir3/package2/__init__.py
		rootDir/emptyDir1/emptyDir2/emptyDir3/package3/__init__.py

	*/
	emptyDir1 := createEmptyDirectory(rootDir, "emptyDir1")
	emptyDir2 := createEmptyDirectory(emptyDir1, "emptyDir2")
	emptyDir3 := createEmptyDirectory(emptyDir2, "emptyDir3")
	pkg1 := createDirectoryWithInitPy(emptyDir1, "package1")
	createDirectoryWithInitPy(emptyDir3, "package2")
	package3 := createDirectoryWithInitPy(emptyDir3, "package3")
	shouldnotpackage31 := createDirectoryWithInitPy(package3, "shouldnotpackage31")
	createDirectoryWithInitPy(shouldnotpackage31, "shouldnotpackage32")
	createDirectoryWithInitPy(pkg1, "shouldnotpackage1")
	createEmptyDirectory(rootDir, "emptyDir")

	// Call the FindPyModules function
	packageNames, err := FindTopLevelModules(rootDir)
	if err != nil {
		t.Fatalf("Error while finding Python modules: %v", err)
	}

	// Check the expected package names
	expectedPackageNames := []string{"package1", "package2", "package3"}
	assert.Equal(t, len(expectedPackageNames), len(packageNames))
	for _, expectedName := range expectedPackageNames {
		found := false
		for packageName, _ := range packageNames {
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

func TestRelativePath(t *testing.T) {
	tests := []struct {
		basePath   string
		fullPath   string
		expected   string
		shouldFail bool
	}{
		{
			basePath:   "/root",
			fullPath:   "/root/subroot/subdir/file.txt",
			expected:   "subroot/subdir/file.txt",
			shouldFail: false,
		},
		{
			basePath:   "/root",
			fullPath:   "/root/file.txt",
			expected:   "file.txt",
			shouldFail: false,
		},
		{
			basePath:   "/root",
			fullPath:   "/anotherroot/file.txt",
			expected:   "",
			shouldFail: true,
		},
		{
			basePath:   "/root",
			fullPath:   "/root",
			expected:   "",
			shouldFail: false,
		},
	}

	for _, test := range tests {
		relativePath, err := RelativePath(test.basePath, test.fullPath)

		if test.shouldFail {
			if err == nil {
				t.Errorf("Expected an error, but got nil")
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if relativePath != test.expected {
				t.Errorf("For BasePath %s and FullPath %s, expected %s but got %s", test.basePath, test.fullPath, test.expected, relativePath)
			}
		}
	}
}

func createTempDirectory(t *testing.T) string {
	tempDir := t.TempDir()
	return tempDir
}

func createDirectoryWithInitPy(rootDir, packageName string) string {
	packageDir := filepath.Join(rootDir, packageName)
	if err := os.Mkdir(packageDir, os.ModePerm); err != nil {
		panic(err)
	}

	initPyFile := filepath.Join(packageDir, "__init__.py")
	if _, err := os.Create(initPyFile); err != nil {
		panic(err)
	}

	return packageDir
}

func createEmptyDirectory(rootDir, directoryName string) string {
	directoryPath := filepath.Join(rootDir, directoryName)
	if err := os.Mkdir(directoryPath, os.ModePerm); err != nil {
		panic(err)
	}

	return directoryPath
}
