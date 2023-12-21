
CODEX is a Code Parser to find imported libraries and building call graph

### Usage
To use the `find-direct-deps` command, run:

```bash
go run main.go scan find-direct-deps --input <project_path>
```

This command scans the specified project path and identifies all the direct dependencies by analyzing the imported modules.

### Example of Imported and Exported Modules

When you run the `find-direct-deps` command, it identifies various imported and exported modules. For instance:

```
**Imported Modules:**
- json
- argparse
- feedparser
- mock
- shlex
- unittest
- git
- titlecase
- nc
- configparser

**Exported Modules:**
- my_project
```

These modules represent the dependencies and outputs that the tool can detect and handle within your project.


## Features 

* Advanced Parsing: Utilizes tree_sitter for syntactic analysis of Python code, enabling accurate identification of both imported and exported modules based on code structure, not just package files.

* Dependency Identification: Differentiates between direct and indirect dependencies by analyzing actual module imports in the code, providing a comprehensive view of the project's dependency structure.

## Import it as Library

### Import it as a module
```
import (
	"github.com/safedep/codex/pkg/parser/py/imports"
	"github.com/safedep/codex/pkg/utils/py/dir"
)

```

### Create Parser

```
cf := imports.NewPyCodeParserFactory()
parser, err := cf.NewCodeParser()
if err != nil {
    logger.Warnf("Error while creating parser %v", err)
    return nil, err
}
```

### Find imported and exported modules
```
/*
Find all imported modules based on the import statements in the code
*/
	ctx := context.Background()
	includeExtensions := []string{".py"}
	excludeDirs := []string{".git", "test"}

	rootPkgs, _ := r.pyCodeParser.FindImportedModules(ctx, sourcePath,
		true, includeExtensions, excludeDirs)
	
    rootPkgs.GetPackagesNames()

/*
Find all exported modules by package itself that can be imported by others
*/
	ctx := context.Background()
	logger.Debugf("Finding Exported module at %s", sourcePath)
	exportedModules, err := r.pyCodeParser.FindExportedModules(ctx, sourcePath)
	if err != nil {
		logger.Debugf("Error while finding exported modules %s", err)
		return nil, err
	}
	modules := exportedModules.GetExportedModules()
	rp, _ := dir.FindTopLevelModules(sourcePath)
	logger.Debugf("Found Exported modules  %s %s %s", modules, rp, err)
}

```

## Roadmap

* Multi Language Support - Java, PHP, NPM









