/*
	Provide methods to get code snippets from a given code file
*/

package imports

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/safedep/codex/pkg/parser/py/utils/dir"
	"github.com/safedep/dry/log"
	sitter "github.com/smacker/go-tree-sitter"
	tree_sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
)

const IMPORT_QUERY = `
(import_statement 
	name: (dotted_name 
			(identifier) @module_name
	)
)

(import_from_statement
	 module_name: (dotted_name) @module_name
	 name: (dotted_name 
			(identifier) @submodule_name
	)
)

(import_from_statement
	 module_name: (relative_import) @module_name
	 name: (dotted_name 
			(identifier) @submodule_name 
	)
)

(import_from_statement
	 module_name: (dotted_name) @module_name
	 name: (aliased_import
			 name: (dotted_name 
				(identifier) @submodule_name
				)
			 alias: (identifier) @submodule_alias
			 
	)
)

(import_from_statement
	 module_name: (relative_import) @module_name
	 name: (aliased_import
			 name: (dotted_name 
				(identifier) @submodule_name
				)
			 alias: (identifier) @submodule_alias
			 
	)
)	
`

type TypedValue struct {
	T        string
	V        string
	RowStart uint32
	RowEnd   uint32
}

type ImportedModule struct {
	Name       TypedValue
	Alias      *TypedValue
	Definition *TypedValue // it can be module, and other definitions
}

type FileCodeAnalysis struct {
	Path    string
	Modules []*ImportedModule
}

type RepoCodeAnalysis struct {
	Path          string
	FilesAnalysis []*FileCodeAnalysis
}

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

func (cpf *CodeParser) FindDirectDependencies(ctx context.Context,
	dirpath string, failOnFirstError bool, includeExtensions, excludeDirs []string) (map[string]string, error) {
	// Start
	rootPackages, _ := dir.FindTopLevelModules(dirpath)
	_, err := cpf.findModulesRecursive(ctx, dirpath, failOnFirstError, includeExtensions, excludeDirs)
	if err != nil {
		return rootPackages, nil
	}
	// log.Debugf("%s", repoAnalysis)
	return rootPackages, nil
}

func (cpf *CodeParser) findModulesRecursive(ctx context.Context,
	rootDir string, failOnFirstError bool, includeExtensions, excludeDirs []string) (*RepoCodeAnalysis, error) {
	// Find modules recursively
	repoAnalysis := &RepoCodeAnalysis{Path: rootDir}
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the directory should be excluded
		if info.IsDir() && cpf.shouldExcludeDir(path, excludeDirs) {
			log.Debugf("Skipping directory .. %s", path)
			return filepath.SkipDir
		}

		// Check if the file should be included based on its extension
		if !info.IsDir() && cpf.shouldIncludeFile(path, includeExtensions) {
			log.Debugf("Parsing file .. %s", path)
			fa, err := cpf.findModulesInFile(ctx, rootDir, path)
			if err != nil {
				log.Debugf("Error while parsing the file %s", path)
				if failOnFirstError {
					return err
				}
				return nil
			}
			repoAnalysis.FilesAnalysis = append(repoAnalysis.FilesAnalysis, fa)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return repoAnalysis, nil
}

// Helper function to check if a directory should be excluded
func (cpf *CodeParser) shouldExcludeDir(dirPath string, excludeDirs []string) bool {
	for _, excludeDir := range excludeDirs {
		if dirPath == excludeDir {
			return true
		}
	}
	return false
}

// Helper function to check if a file should be included based on its extension
func (cpf *CodeParser) shouldIncludeFile(filePath string, includeExtensions []string) bool {
	ext := filepath.Ext(filePath)
	for _, includeExt := range includeExtensions {
		if ext == includeExt {
			return true
		}
	}
	return false
}

func (cpf *CodeParser) findModulesInFile(ctx context.Context,
	rootDir string, filepath string) (*FileCodeAnalysis, error) {

	parsedCode, err := cpf.ParseFile(ctx, filepath)
	if err != nil {
		log.Debugf("Error while parsing file to parsed code")
		return nil, err
	}

	modules, err := parsedCode.ExtractModules()
	if err != nil {
		log.Debugf("Error while extracting modules from the file %s", filepath)
		return nil, err
	}
	path, err := dir.RelativePath(rootDir, filepath)
	if err != nil {
		log.Debugf("Error while extracting relative path %s %s", rootDir, filepath)
		return nil, err
	}
	fca := &FileCodeAnalysis{Modules: modules, Path: path}
	return fca, nil

}

// ParseCode reads and parses code from the specified file path using a PyCodeParserFactory.
func (cpf *CodeParser) ParseFile(ctx context.Context, filepath string) (*ParsedCode, error) {
	// Open the file using os.Open method instead of ioutil.ReadFile
	file, err := os.Open(filepath)
	if err != nil {
		log.Debugf("Error opening file: %v", err)
		return nil, err
	}
	defer file.Close() // Close the file when the function exits

	// Get the file size to create a buffer of appropriate size
	fileInfo, err := file.Stat()
	if err != nil {
		log.Debugf("Error getting file info: %v", err)
		return nil, err
	}

	// Read the file content into a buffer
	code := make([]byte, fileInfo.Size())
	_, err = file.Read(code)
	if err != nil {
		log.Debugf("Error reading file: %v", err)
		return nil, err
	}

	// Parse the code using the created parser
	parsedCode, err := cpf.parseCode(ctx, nil, code)
	if err != nil {
		log.Warnf("Error while parsing code: %v", err)
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
		log.Debugf("Error while parsing code %v", err)
		return nil, err
	}

	if tree.RootNode() == nil {
		return nil, fmt.Errorf("Error parsing code. Found nil root node")
	}
	return &ParsedCode{codeTree: tree, code: content, lang: cp.lang}, nil
}

func (s *ParsedCode) ExtractModules() ([]*ImportedModule, error) {
	// Parse source code
	lang := s.lang
	// Execute the query
	q, err := sitter.NewQuery([]byte(IMPORT_QUERY), lang)
	modules := make([]*ImportedModule, 0)
	if err != nil {
		fmt.Println(err)
		return modules, err
	}
	qc := tree_sitter.NewQueryCursor()
	qc.Exec(q, s.codeTree.RootNode())
	// Iterate over query results
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		// Apply predicates filtering
		m = qc.FilterPredicates(m, s.code)
		mod := ImportedModule{}
		var value TypedValue
		for i, c := range m.Captures {
			value = TypedValue{T: c.Node.Type(), V: c.Node.Content(s.code),
				RowStart: c.Node.StartPoint().Row,
				RowEnd:   c.Node.EndPoint().Row}

			// fmt.Printf("%d, %s %s\n", i, c.Node.Type(), c.Node.Content(s.code))

			if i == 0 {
				// Module Name
				mod.Name = value
			} else if i == 1 {
				// Any definitions
				mod.Definition = &value
			} else if i == 2 {
				// Any alias
				mod.Alias = &value
			}

			modules = append(modules, &mod)
		}
	}

	return modules, nil
}
