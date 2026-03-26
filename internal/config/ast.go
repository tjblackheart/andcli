package config

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

func apply(af *ast.File, patch map[string]any) error {
	for pathStr, value := range patch {
		if err := replace(af, pathStr, value); err != nil {
			return err
		}
	}
	return nil
}

func parse(path string) (*ast.File, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parser.ParseBytes(b, parser.ParseComments)
}

func replace(af *ast.File, pathStr string, value any) error {
	path, err := yaml.PathString(pathStr)
	if err != nil {
		return err
	}

	node, err := path.FilterFile(af)
	if err != nil {
		return nil
	}

	var newNode ast.Node
	switch v := value.(type) {
	case bool:
		newNode = &ast.BoolNode{
			BaseNode: &ast.BaseNode{},
			Token:    node.GetToken(),
			Value:    v,
		}
	default:
		newNode = &ast.StringNode{
			BaseNode: &ast.BaseNode{},
			Token:    node.GetToken(),
			Value:    fmt.Sprint(v),
		}
	}

	if comment := node.GetComment(); comment != nil {
		newNode.SetComment(comment)
	}

	return path.ReplaceWithNode(af, newNode)
}
