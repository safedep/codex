package imports

import (
	"io/ioutil"
	"testing"

	tree_sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
)

// MockCodeParserFactory implements CodeParserFactory for testing purposes.
type MockCodeParserFactory struct{}

func (mcpf *MockCodeParserFactory) NewCodeParser() (*CodeParser, error) {
	lang := python.GetLanguage()
	parser := tree_sitter.NewParser()
	parser.SetLanguage(lang)
	codeParser := &CodeParser{parser: parser, lang: lang}
	return codeParser, nil
}

// Helper function to create a temporary test file with Python code
func createTempPythonFile(t *testing.T, code string) string {
	tempFile, err := ioutil.TempFile("", "test_python_code_*.py")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}

	_, err = tempFile.WriteString(code)
	if err != nil {
		t.Fatalf("Error writing to temp file: %v", err)
	}

	// Close the tempFile to release resources
	tempFile.Close()

	return tempFile.Name()
}
