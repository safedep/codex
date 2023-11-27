package imports

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeMethodMap(t *testing.T) {
	// Create a ParsedCode instance with a known source code
	sourceCode := `
class MyClass:
    def method1(self):
        pass

    @staticmethod
    def static_method():
        pass

class AnotherClass:
    def method2(self):
        pass
`

	tempFile := createTempPythonFile(t, sourceCode)
	defer os.ReadFile(tempFile)

	// Create a CodeParserFactory instance
	cpf := &MockCodeParserFactory{}

	// Create a CodeParser using the factory
	codeParser, err := cpf.NewCodeParser()
	if err != nil {
		t.Fatalf("Error creating CodeParser: %v", err)
	}

	// Parse the test file
	ctx := context.TODO()
	rootDir, relFilePath := path.Split(tempFile)
	parsedCode, err := codeParser.ParseFile(ctx, rootDir, relFilePath)
	if err != nil {
		t.Fatalf("Error parsing file: %v", err)
	}

	// Call MakeMethodMap on the parsedCode
	methodMap, err := parsedCode.MakeMethodMap()
	assert.NoError(t, err)

	// Assert expected values for specific methods
	expectedMethods := map[MethodMapKey]MethodInfo{
		{path: relFilePath, className: "MyClass", name: "method1"}: {
			index:         0,
			node:          nil, // Set the expected value based on your test case
			name:          "method1",
			invoketype:    "invokevirtual", // Set the expected value based on your test case
			argumentCount: 1,               // Set the expected value based on your test case
		},
		{path: relFilePath, className: "MyClass", name: "static_method"}: {
			index:         1,
			node:          nil, // Set the expected value based on your test case
			name:          "static_method",
			invoketype:    "invokestatic", // Set the expected value based on your test case
			argumentCount: 0,              // Set the expected value based on your test case
		},
		{path: relFilePath, className: "AnotherClass", name: "method2"}: {
			index:         2,
			node:          nil, // Set the expected value based on your test case
			name:          "method2",
			invoketype:    "invokevirtual", // Set the expected value based on your test case
			argumentCount: 1,               // Set the expected value based on your test case
		},
	}

	for key, expectedInfo := range expectedMethods {
		actualInfo, ok := methodMap.methods[key]
		assert.True(t, ok)
		assert.Equal(t, expectedInfo.name, actualInfo.name)
		assert.Equal(t, expectedInfo.invoketype, actualInfo.invoketype)
		assert.Equal(t, expectedInfo.argumentCount, actualInfo.argumentCount)
		assert.Equal(t, expectedInfo.index, actualInfo.index)
		assert.NotNil(t, actualInfo.node)
	}
}
