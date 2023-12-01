/*
	Provide methods to get code snippets from a given code file
*/

package imports

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/safedep/codex/pkg/utils/py/dir"
	"github.com/safedep/dry/log"
	tree_sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
)

const FUNC_DEFINITION_QUERY = `
(function_definition
	name: (identifier) @method_name
	parameters: (parameters) @method_params) @method
`

const IMPORT_QUERY = `
(import_statement 
	name: ((dotted_name) @module_name
	)
)

(import_from_statement
	 module_name: (dotted_name) @module_name
	 name: (dotted_name 
			(identifier) @submodule_name @submodule_alias
	)
)

(import_from_statement
	 module_name: (relative_import) @module_name
	 name: (dotted_name 
			(identifier) @submodule_name @submodule_alias
	)
)

(import_statement 
	 name: (aliased_import
			 name: (
				 (dotted_name) @module_name @submodule_name
			 )
			 alias: (identifier) @submodule_alias
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
			 name: (
				 (dotted_name) @submodule_name
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

type ImportedModules struct {
	pkgNames map[string]bool
}

func NewImportedModules() *ImportedModules {
	return &ImportedModules{pkgNames: make(map[string]bool, 0)}
}

func (dd *ImportedModules) addDependency(pkg string, path string) {
	dd.pkgNames[pkg] = true
}

func (dd *ImportedModules) GetPackagesNames() []string {
	pkgs := make([]string, 0)
	for pkg, _ := range dd.pkgNames {
		pkgs = append(pkgs, pkg)
	}

	return pkgs
}

type ExportedModules struct {
	pkgNames map[string]string
}

func NewExportedModules() *ExportedModules {
	return &ExportedModules{pkgNames: make(map[string]string, 0)}
}

func (dd *ExportedModules) addModule(pkg string, path string) {
	dd.pkgNames[pkg] = path
}

func (dd *ExportedModules) GetExportedModules() []string {
	pkgs := make([]string, 0)
	for pkg, _ := range dd.pkgNames {
		pkgs = append(pkgs, pkg)
	}

	return pkgs
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
	path     string // file path of the file
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

// FindImportedModules analyzes the code repository in the specified directory and returns its direct dependencies.
func (cpf *CodeParser) FindImportedModules(ctx context.Context,
	dirpath string, failOnFirstError bool,
	includeExtensions, excludeDirs []string) (*ImportedModules, error) {
	// Find top-level modules in the provided directory.
	rootPackages, _ := dir.FindTopLevelModules(dirpath)

	// Find modules recursively in the directory, filtering based on file extensions and excluding specified directories.
	repoAnalysis, err := cpf.findModulesRecursive(ctx, dirpath, failOnFirstError, includeExtensions, excludeDirs)
	if err != nil {
		// If there is an error during analysis, return an error.
		return nil, nil
	}

	// Find unique modules/packages in the analyzed code files.
	dd := cpf.findUniqueModules(rootPackages, repoAnalysis)

	// Return the direct dependencies found.
	return dd, nil
}

// FindImportedModules analyzes the code repository in the specified directory and returns its direct dependencies.
func (cpf *CodeParser) FindExportedModules(ctx context.Context,
	dirpath string) (*ExportedModules, error) {
	// Find top-level modules in the provided directory.
	rootPackages, _ := dir.FindTopLevelModules(dirpath)
	exportedModules := NewExportedModules()
	for name, path := range rootPackages {
		exportedModules.addModule(name, path)
	}
	// Return the direct dependencies found.
	return exportedModules, nil
}

// findUniqueModules finds and returns unique modules/packages from the analyzed code files.
func (cpf *CodeParser) findUniqueModules(rootPackages map[string]string,
	repoAnalysis *RepoCodeAnalysis) *ImportedModules {
	// Create a new ImportedModules instance to store the results.
	dd := NewImportedModules()

	// Create a map to track unique module/package names.
	uniqueModNames := map[string]bool{}

	// Iterate through the analyzed code files.
	for _, fa := range repoAnalysis.FilesAnalysis {
		for _, mod := range fa.Modules {
			// Initialize flags to check for relative and local imports.
			isRelativeImport := false
			isImportLocal := false

			// Iterate through the root packages to check for relative and local imports.
			for pkg := range rootPackages {
				isRelativeImport = isRelativeImport || strings.HasPrefix(mod.Name.V, ".")
				isImportLocal = isImportLocal || strings.HasPrefix(mod.Name.V, fmt.Sprintf("%s.", pkg)) || mod.Name.V == pkg
			}

			// If it's neither a relative nor local import, consider it a direct dependency.
			if !isRelativeImport && !isImportLocal {
				// Extract the top-level package name.
				topLevelPkg := dir.SplitAndGetLeftMost(mod.Name.V, ".")
				// Add the top-level package as a direct dependency.
				dd.addDependency(topLevelPkg, fa.Path)
				// Mark the package name as unique.
				uniqueModNames[topLevelPkg] = true
			}
		}
	}

	// Return the ImportedModules instance containing the direct dependencies.
	return dd
}

// findModulesRecursive recursively analyzes code files in a directory.
func (cpf *CodeParser) findModulesRecursive(ctx context.Context,
	rootDir string, failOnFirstError bool, includeExtensions, excludeDirs []string) (*RepoCodeAnalysis, error) {
	// Create a RepoCodeAnalysis instance to store the analysis results.
	repoAnalysis := &RepoCodeAnalysis{Path: rootDir}

	// Walk through the directory tree to analyze files and subdirectories.
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the directory should be excluded based on the provided exclusion list.
		if info.IsDir() && cpf.shouldExcludeDir(path, excludeDirs) {
			log.Debugf("Skipping directory .. %s", path)
			return filepath.SkipDir
		}

		// Check if the file should be included based on its extension and analyze it if so.
		if !info.IsDir() && cpf.shouldIncludeFile(path, includeExtensions) {
			relPath, err := dir.RelativePath(rootDir, path)
			if err != nil {
				log.Debugf("Error while getting relative path %s", err)
				return err
			}
			fa, err := cpf.findModulesInFile(ctx, rootDir, relPath)
			if err != nil {
				log.Debugf("Error while parsing the file %s", path)
				if failOnFirstError {
					// If failOnFirstError is true, return the error immediately.
					return err
				}
				// If failOnFirstError is false, continue analyzing other files.
				return nil
			}
			// Append the file analysis results to the repository analysis.
			repoAnalysis.FilesAnalysis = append(repoAnalysis.FilesAnalysis, fa)
		}

		return nil
	})

	if err != nil {
		// If there is an error during analysis, return the error.
		return nil, err
	}

	// Return the RepoCodeAnalysis instance containing the analysis results.
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
	rootDir string, relFilePath string) (*FileCodeAnalysis, error) {

	parsedCode, err := cpf.ParseFile(ctx, rootDir, relFilePath)
	if err != nil {
		log.Debugf("Error while parsing file to parsed code")
		return nil, err
	}

	modules, err := parsedCode.ExtractModules()
	if err != nil {
		log.Debugf("Error while extracting modules from the file %s %s", rootDir, relFilePath)
		return nil, err
	}

	fca := &FileCodeAnalysis{Modules: modules, Path: relFilePath}
	return fca, nil

}

// ParseCode reads and parses code from the specified file path using a PyCodeParserFactory.
func (cpf *CodeParser) ParseFile(ctx context.Context, rootDir string, relFilePath string) (*ParsedCode, error) {
	filepath := path.Join(rootDir, relFilePath)
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
	parsedCode, err := cpf.parseCode(ctx, nil, code, relFilePath)
	if err != nil {
		log.Warnf("Error while parsing code: %v", err)
		return nil, err
	}

	// Return the parsed code
	return parsedCode, nil
}

func (cp *CodeParser) parseCode(ctx context.Context,
	parentTree *tree_sitter.Tree,
	content []byte,
	sourcePath string) (*ParsedCode, error) {
	tree, err := cp.parser.ParseCtx(ctx, parentTree, content)
	if err != nil {
		log.Debugf("Error while parsing code %v", err)
		return nil, err
	}

	if tree.RootNode() == nil {
		return nil, fmt.Errorf("Error parsing code. Found nil root node")
	}
	return &ParsedCode{codeTree: tree, code: content,
		lang: cp.lang, path: sourcePath}, nil
}

func (s *ParsedCode) ExtractModules() ([]*ImportedModule, error) {
	// Parse source code
	lang := s.lang
	// Execute the query
	q, err := tree_sitter.NewQuery([]byte(IMPORT_QUERY), lang)
	modules := make([]*ImportedModule, 0)
	if err != nil {
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
