package ts

import (
	tree_sitter "github.com/smacker/go-tree-sitter"
)

func FindClosestAncestorOfType(root *tree_sitter.Node, ancestorType string) *tree_sitter.Node {
	root = root.Parent() // Skip over root itself since root is not an ancestor candidate

	for root != nil && root.Type() != ancestorType {
		root = root.Parent()
		if root == nil {
			return nil
		}
	}

	return root
}

func FindAllAncestorsOfTypes(root *tree_sitter.Node, ancestorTypes []string) []*tree_sitter.Node {
	var ancestors []*tree_sitter.Node

	for root != nil {
		root = root.Parent() // Note that root itself is skipped
		if root != nil {
			for _, ancestorType := range ancestorTypes {
				if root.Type() == ancestorType {
					ancestors = append(ancestors, root)
				}
			}
		}
	}

	return ancestors
}

const defaultUnused = -999

func FindFirstChildOfType(root *tree_sitter.Node, childType string, level int) *tree_sitter.Node {
	assertLevelUsed := level != defaultUnused

	var recur func(root *tree_sitter.Node, level int) *tree_sitter.Node
	recur = func(root *tree_sitter.Node, level int) *tree_sitter.Node {
		if root == nil {
			return nil
		}
		if assertLevelUsed && level < 0 {
			return nil
		}
		if assertLevelUsed && level == 0 {
			if root.Type() == childType {
				return root
			}
			return nil
		}

		for i := uint32(0); i < root.ChildCount(); i++ {
			child := root.Child(int(i))
			if child.Type() == childType {
				return child
			}
			res := recur(child, level-1)
			if res != nil {
				return res
			}
		}

		return nil
	}

	return recur(root, level)
}

func FindAllChildrenOfType(root *tree_sitter.Node, childType string, level int) []*tree_sitter.Node {
	if root == nil {
		return nil
	}

	assertLevelUsed := level != defaultUnused

	var res []*tree_sitter.Node

	queue := []*tree_sitter.Node{root}

	for !assertLevelUsed || level >= 0 {
		numNodes := len(queue)
		for i := 0; i < numNodes; i++ {
			top := queue[0]
			queue = queue[1:]
			if top.Type() == childType {
				res = append(res, top)
			}
			for i := uint32(0); i < top.ChildCount(); i++ {
				child := root.Child(int(i))
				queue = append(queue, child)
			}
		}
		if len(queue) == 0 || (assertLevelUsed && level < 0) {
			return res
		}
		level--
	}

	return res
}
