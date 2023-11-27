package imports

import (
	"fmt"

	"github.com/safedep/codex/pkg/utils/ts"
	tree_sitter "github.com/smacker/go-tree-sitter"
)

type MethodInfo struct {
	index         int
	node          *tree_sitter.Node // Assuming you have a Node struct defined
	name          string
	invoketype    InvokeType // Assuming 'invstr' is a string
	argumentCount int        // Assuming 'method_params_node.named_child_count' is an int
}

type MethodMapKey struct {
	path      string // source file path
	name      string // method name
	className string // if any
}

type MethodMap struct {
	methods map[MethodMapKey]MethodInfo
}

func (s *ParsedCode) MakeMethodMap() (*MethodMap, error) {
	methodDict := make(map[MethodMapKey]MethodInfo)
	methodIndex := 0
	// Parse source code
	lang := s.lang
	// Execute the query
	q, err := tree_sitter.NewQuery([]byte(FUNC_DEFINITION_QUERY), lang)
	if err != nil {
		return nil, err
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
		c := m.Captures[0] // only the method capture
		methodNode := c.Node
		classDecNode := ts.FindClosestAncestorOfType(methodNode, "class_definition")
		classNodeName := ts.FindFirstChildOfType(classDecNode, "identifier", 1)
		methodNodeName := ts.FindFirstChildOfType(methodNode, "identifier", 1)
		descriptor_name_node := ts.FindClosestAncestorOfType(methodNode, "decorated_definition")
		descriptor_name_node = ts.FindFirstChildOfType(descriptor_name_node, "decorator", 1)
		method_params_node := ts.FindFirstChildOfType(methodNode, "parameters", 1)

		className := s.getContentIfNotNil(classNodeName)
		methodName := s.getContentIfNotNil(methodNodeName)
		descName := s.getContentIfNotNil(descriptor_name_node)
		methodParams := s.getContentIfNotNil(method_params_node)

		invokeType := getInvokeType(descName, className)
		key := MethodMapKey{path: s.path, className: className, name: methodName}
		methodInfo := MethodInfo{index: methodIndex,
			node:          methodNode,
			name:          methodName,
			invoketype:    invokeType,
			argumentCount: int(method_params_node.ChildCount())}

		methodDict[key] = methodInfo
		methodIndex += 1

		fmt.Println(className, methodName, descName, methodParams)
	}

	methods := &MethodMap{methods: methodDict}
	return methods, nil
}

func (s *ParsedCode) getContentIfNotNil(node *tree_sitter.Node) string {
	if node != nil {
		return node.Content(s.code)
	} else {
		return ""
	}
}
