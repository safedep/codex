/*
	Provide methods to get code snippets from a given code file
*/

package parser

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/safedep/vet/pkg/common/logger"
	sitter "github.com/smacker/go-tree-sitter"
	tree_sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
)

const (
	node_type_identifier          = "identifier"
	node_type_function_definition = "function_definition"
	node_type_string              = "string"
	node_type_assignment          = "assignment"
	node_type_format_string       = "format_string"
	node_type_expr_list           = "expr_list"
	node_type_testlist            = "testlist"
	node_type_keyword_argument    = "keyword_argument"
	node_type_call                = "call"
	node_type_def                 = "def"
)

type CodeSnippetFactory struct {
}

type CodeBlockResult struct {
	Code             string
	BlockIndentation string
}

func (cp *CodeBlockResult) GetIndentedCode() string {
	return cp.BlockIndentation + cp.Code
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

func NewCodeSnippetFactory() *CodeSnippetFactory {
	return &CodeSnippetFactory{}
}

// ParseCode reads and parses code from the specified file path using a CodeSnippetFactory.
func (cpf *CodeSnippetFactory) ParseCode(ctx context.Context, filepath string) (*ParsedCode, error) {
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

	// TODO: Detect the language. Currently, assuming it is Python code.
	lang := python.GetLanguage()
	parser, err := NewCodeParser(lang)
	if err != nil {
		logger.Warnf("Error while creating parser %v", err)
		return nil, err
	}

	// Parse the code using the created parser
	parsedCode, err := parser.Parse(ctx, nil, code)
	if err != nil {
		logger.Warnf("Error while parsing code: %v", err)
		return nil, err
	}

	// Return the parsed code
	return parsedCode, nil
}

func NewCodeParser(lang *tree_sitter.Language) (*CodeParser, error) {
	parser := tree_sitter.NewParser()
	parser.SetLanguage(lang)
	codeParser := &CodeParser{parser: parser, lang: lang}
	return codeParser, nil
}

func (cp *CodeParser) Parse(ctx context.Context,
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

func (s *ParsedCode) Analyze() {
	s.extractStringConstants(s.codeTree.RootNode(), s.code)
}

// extractStringConstants recursively extracts string constants and symbols from the syntax tree nodes.
func (s *ParsedCode) extractStringConstants(node *tree_sitter.Node, code []byte) ([]string, []string) {
	const_strings := []string{}
	symbol_string := []string{}
	fmt.Println(node.Type(), node)
	switch node.Type() {
	case node_type_identifier:
		for i := uint32(0); i < node.ChildCount(); i++ {
			child := node.Child(int(i))
			a, b := s.extractStringConstants(child, code)
			const_strings = append(const_strings, a...)
			symbol_string = append(symbol_string, b...)
		}
		symbol_string = append(symbol_string, node.Content(code))
	case node_type_function_definition:
		if node.Child(0).Type() == node_type_def && node.Child(1).Type() == node_type_identifier {
			const_strings2 := make([]string, 0)
			symbol_strings2 := []string{}
			for i := uint32(2); i < node.ChildCount(); i++ {
				child := node.Child(int(i))
				a, b := s.extractStringConstants(child, code)
				const_strings2 = append(const_strings2, a...)
				symbol_strings2 = append(symbol_strings2, b...)
			}
			const_strings = append(const_strings, const_strings2...)
			symbol_string = append(symbol_string, symbol_strings2...)
			symbol_string = append(symbol_string, node.Child(1).Content(code))
			// s.Symbol2strings[node.Child(1).Content(code)] = const_strings2
			// s.Symbol2symbols[node.Child(1).Content(code)] = symbol_strings2

		}
	case node_type_string:
		a := strings.Trim(string(node.Content(code)), "\"")
		const_strings = append(const_strings, a)
	case node_type_assignment:
		attribute := node.Child(0)
		if attribute.Type() == node_type_identifier {
			const_strings2 := make([]string, 0)
			symbol_strings2 := []string{}
			for i := uint32(1); i < node.ChildCount(); i++ {
				child := node.Child(int(i))
				a, b := s.extractStringConstants(child, code)
				const_strings2 = append(const_strings2, a...)
				symbol_strings2 = append(symbol_strings2, b...)
			}
			const_strings = append(const_strings, const_strings2...)
			symbol_string = append(symbol_string, symbol_strings2...)
			symbol_string = append(symbol_string, attribute.Content(code))
			// s.Symbol2strings[attribute.Content(code)] = const_strings2
			// s.Symbol2symbols[attribute.Content(code)] = symbol_strings2

		}
	case node_type_format_string:
		a, b := s.extractStringConstants(node.Child(0), code)
		const_strings = append(const_strings, a...)
		symbol_string = append(symbol_string, b...)
	case node_type_expr_list, node_type_testlist:
		for i := uint32(0); i < node.ChildCount(); i++ {
			child := node.Child(int(i))
			a, b := s.extractStringConstants(child, code)
			const_strings = append(const_strings, a...)
			symbol_string = append(symbol_string, b...)
		}
	case node_type_keyword_argument:
		{
			attribute := node.Child(0)
			if attribute.Type() == node_type_identifier {
				const_strings2 := make([]string, 0)
				symbol_strings2 := make([]string, 0)
				for i := uint32(1); i < node.ChildCount(); i++ {
					child := node.Child(int(i))
					a, b := s.extractStringConstants(child, code)
					const_strings2 = append(const_strings2, a...)
					symbol_strings2 = append(symbol_strings2, b...)
				}
				const_strings = append(const_strings, const_strings2...)
				symbol_string = append(symbol_string, symbol_strings2...)

				// s.Symbol2strings[attribute.Content(code)] = const_strings2
				// s.Symbol2symbols[attribute.Content(code)] = symbol_strings2
			}
		}
	case node_type_call:
		for i := uint32(0); i < node.ChildCount(); i++ {
			child := node.Child(int(i))
			a, b := s.extractStringConstants(child, code)
			const_strings = append(const_strings, a...)
			symbol_string = append(symbol_string, b...)
		}
	default:
		for i := uint32(0); i < node.ChildCount(); i++ {
			child := node.Child(int(i))
			a, b := s.extractStringConstants(child, code)
			const_strings = append(const_strings, a...)
			symbol_string = append(symbol_string, b...)
		}
	}

	return const_strings, symbol_string
}

func (pc *ParsedCode) GetCodeBlock(lineNumber uint32) (*CodeBlockResult, error) {
	rootNode := pc.codeTree.RootNode()
	node := pc.findNodeAtLineNumber(rootNode, lineNumber)

	if node != nil {
		topLevelBlock := pc.findTopLevelCodeBlock(pc.code, node)
		// fmt.Printf("Top level code block: %s %d %d \n", node.Type(), node.StartPoint().Row, node.EndPoint().Row)
		// fmt.Println(topLevelBlock.BlockIndentation + topLevelBlock.Code)
		// fmt.Printf("Indentation%s###\n", topLevelBlock.BlockIndentation)
		return topLevelBlock, nil
	} else {
		return nil, fmt.Errorf("Line number not found in the syntax tree")
	}
}

func (pc *ParsedCode) findNodeAtLineNumber(node *tree_sitter.Node, lineNumber uint32) *tree_sitter.Node {
	var deepestDefNode *tree_sitter.Node

	if node.StartPoint().Row <= lineNumber && lineNumber <= node.EndPoint().Row {
		// fmt.Printf("Checking node %s Row at line %d %d\n", node.Type(), node.StartPoint().Row, node.EndPoint().Row)
		if node.Type() == "function_definition" {
			deepestDefNode = node
		}

		if node.ChildCount() > 0 {
			for i := 0; uint32(i) < node.ChildCount(); i++ {
				child := node.Child(i)
				childNode := pc.findNodeAtLineNumber(child, lineNumber)
				if childNode != nil && childNode.Type() == "function_definition" {
					deepestDefNode = childNode
				}
			}
		}
	}
	return deepestDefNode
}

func (pc *ParsedCode) findTopLevelCodeBlock(code []byte, node *tree_sitter.Node) *CodeBlockResult {
	startIndex := int(node.StartByte())
	endIndex := int(node.EndByte())

	// Ensure startIndex and endIndex are within the bounds of the code slice
	if startIndex >= 0 && endIndex <= len(code) {
		codeBlock := string(code[startIndex:endIndex])

		// Calculate leading spaces or tabs (indentation)
		indentation := ""
		// Scan backward from the startIndex to find leading spaces or tabs
		for i := startIndex - 1; i >= 0; i-- {
			char := code[i]
			if char == ' ' || char == '\t' {
				indentation = string(char) + indentation
			} else {
				break
			}
		}

		return &CodeBlockResult{Code: codeBlock, BlockIndentation: indentation}
	}
	return nil
}
