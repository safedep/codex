/*
	Provide methods to get code snippets from a given code file
*/

package imports

import (
	"context"
	"fmt"
	"os"

	"github.com/safedep/vet/pkg/common/logger"
	sitter "github.com/smacker/go-tree-sitter"
	tree_sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
)

type PyCodeParserFactory struct {
}

type CodeParser struct {
	parser *tree_sitter.Parser
	lang   *tree_sitter.Language
}

type ParsedCode struct {
	codeTree *tree_sitter.Tree
	code     []byte // Original Code Content
	lang     *tree_sitter.Language
}

func NewPyCodeParserFactory() *PyCodeParserFactory {
	return &PyCodeParserFactory{}
}

func (cpf *PyCodeParserFactory) NewCodeParser() (*CodeParser, error) {
	// TODO: Detect the language. Currently, assuming it is Python code.
	lang := python.GetLanguage()
	parser := tree_sitter.NewParser()
	parser.SetLanguage(lang)
	codeParser := &CodeParser{parser: parser, lang: lang}
	return codeParser, nil
}

// ParseCode reads and parses code from the specified file path using a PyCodeParserFactory.
func (cpf *CodeParser) ParseFile(ctx context.Context, filepath string) (*ParsedCode, error) {
	// Open the file using os.Open method instead of ioutil.ReadFile
	file, err := os.Open(filepath)
	if err != nil {
		logger.Debugf("Error opening file: %v", err)
		return nil, err
	}
	defer file.Close() // Close the file when the function exits

	// Get the file size to create a buffer of appropriate size
	fileInfo, err := file.Stat()
	if err != nil {
		logger.Debugf("Error getting file info: %v", err)
		return nil, err
	}

	// Read the file content into a buffer
	code := make([]byte, fileInfo.Size())
	_, err = file.Read(code)
	if err != nil {
		logger.Debugf("Error reading file: %v", err)
		return nil, err
	}

	// Parse the code using the created parser
	parsedCode, err := cpf.parseCode(ctx, nil, code)
	if err != nil {
		logger.Warnf("Error while parsing code: %v", err)
		return nil, err
	}

	// Return the parsed code
	return parsedCode, nil
}

func (cp *CodeParser) parseCode(ctx context.Context,
	parentTree *tree_sitter.Tree,
	content []byte) (*ParsedCode, error) {
	tree, err := cp.parser.ParseCtx(ctx, parentTree, content)
	if err != nil {
		logger.Debugf("Error while parsing code %v", err)
		return nil, err
	}

	if tree.RootNode() == nil {
		return nil, fmt.Errorf("Error parsing code. Found nil root node")
	}
	return &ParsedCode{codeTree: tree, code: content, lang: cp.lang}, nil
}

func (s *ParsedCode) Query(query string) error {
	// Parse source code
	lang := s.lang
	// Execute the query
	q, err := sitter.NewQuery([]byte(query), lang)
	if err != nil {
		fmt.Println(err)
		return err
	}
	qc := tree_sitter.NewQueryCursor()
	qc.Exec(q, s.codeTree.RootNode())
	// Iterate over query results
	fmt.Println("Going to iterate..")
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		// Apply predicates filtering
		m = qc.FilterPredicates(m, s.code)
		for i, c := range m.Captures {
			fmt.Printf("%d, %s %s\n", i, c.Node.Type(), c.Node.Content(s.code))
		}
	}

	return nil
}
