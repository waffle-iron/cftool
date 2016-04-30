package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/commondream/yaml-ast"
)

type tagHandler func(string, string) *yamlast.Node

func getTagHandler(tag string) tagHandler {
	switch tag {
	case "!import":
		return importTagHandler
	case "!ref":
		return refHandler
	default:
		return nil
	}
}

func loadTemplate(path string) *yamlast.Node {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file %s.\n", path)
		os.Exit(1)
	}

	doc := yamlast.Parse(b)
	processTags(doc)
	return doc
}

func processTags(node *yamlast.Node) {
	for index, child := range node.Children {
		if child.Tag != "" {
			handler := getTagHandler(child.Tag)
			if handler != nil {
				node.Children[index] = handler(child.Tag, child.Value)
			}
		}

		processTags(child)
	}
}

func importTagHandler(tag string, value string) *yamlast.Node {
	subDoc := loadTemplate(fmt.Sprintf("./imports/%s.yml", value))
	return subDoc.Children[0]
}

func refHandler(tag string, value string) *yamlast.Node {
	refNode := yamlast.Node{Kind: yamlast.MappingNode}
	refNode.Children = append(refNode.Children,
		&yamlast.Node{Kind: yamlast.ScalarNode, Value: "Ref"})
	refNode.Children = append(refNode.Children,
		&yamlast.Node{Kind: yamlast.ScalarNode, Value: value})

	return &refNode
}

// Converts a template to a json string.
func templateToJSON(node *yamlast.Node) string {
	jsonData, err := json.MarshalIndent(nodeToInterface(node), "", "  ")
	if err != nil {
		panic(err)
	}

	return string(jsonData)
}

// Converts a node to an object
func nodeToInterface(node *yamlast.Node) interface{} {
	switch node.Kind {
	case yamlast.DocumentNode:
		if len(node.Children) > 0 {
			return nodeToInterface(node.Children[0])
		}
		return nil

	case yamlast.MappingNode:
		mapping := map[string]interface{}{}

		for i := 0; i < len(node.Children)/2; i++ {
			key := node.Children[i*2]
			value := node.Children[i*2+1]

			mapping[key.Value] = nodeToInterface(value)
		}
		return mapping

	case yamlast.SequenceNode:
		sequence := []interface{}{}

		for _, child := range node.Children {
			sequence = append(sequence, nodeToInterface(child))
		}

		return sequence
	case yamlast.ScalarNode:
		return node.Value
	case yamlast.AliasNode:
		return node.Value

	default:
		panic("Unsupported node type.")
	}
}
