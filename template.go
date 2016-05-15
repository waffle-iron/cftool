package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/commondream/yaml-ast"
)

type tagHandler func(*Config, string, string) (*yamlast.Node, error)

func getTagHandler(tag string) tagHandler {
	switch tag {
	case "!import":
		return importTagHandler
	case "!ref":
		return refHandler
	case "!file":
		return fileHandler
	case "!vault":
		return vaultHandler
	default:
		return nil
	}
}

func loadTemplate(path string, config *Config) (*yamlast.Node, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error reading file %s: $s", path, err))
	}

	doc, err := yamlast.Parse(b)
	if err != nil {
		return nil, err
	}
	processTree(doc, config)
	return doc, nil
}

func processTree(node *yamlast.Node, config *Config) error {
	for index, child := range node.Children {
		if child.Tag != "" {
			handler := getTagHandler(child.Tag)
			if handler != nil {
				var err error
				node.Children[index], err = handler(config, child.Tag, child.Value)
				if err != nil {
					return err
				}
			}
		}

		err := processTree(child, config)
		if err != nil {
			return err
		}
	}

	return nil
}

func importTagHandler(config *Config, tag string, value string) (*yamlast.Node, error) {
	subDoc, err := loadTemplate(fmt.Sprintf("./imports/%s.yml", value), config)
	if err != nil {
		return nil, err
	}

	return subDoc.Children[0], nil
}

func refHandler(config *Config, tag string, value string) (*yamlast.Node, error) {
	refNode := yamlast.Node{Kind: yamlast.MappingNode}
	refNode.Children = append(refNode.Children,
		&yamlast.Node{Kind: yamlast.ScalarNode, Value: "Ref"})
	refNode.Children = append(refNode.Children,
		&yamlast.Node{Kind: yamlast.ScalarNode, Value: value})

	return &refNode, nil
}

func fileHandler(config *Config, tag string, value string) (*yamlast.Node, error) {
	path := fmt.Sprintf("./files/%s", value)
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error loading file %s: %s", path, err.Error()))
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text()+"\n")
	}
	if scanner.Err() != nil {
		return nil, errors.New(fmt.Sprintf("Error reading file %s: %s", path, scanner.Err().Error()))
	}

	fileNode := yamlast.Node{Kind: yamlast.SequenceNode}
	for _, line := range lines {
		fileNode.Children = append(fileNode.Children,
			&yamlast.Node{Kind: yamlast.ScalarNode, Value: line})
	}

	joinArgs := yamlast.Node{Kind: yamlast.SequenceNode}
	joinArgs.Children = append(joinArgs.Children,
		&yamlast.Node{Kind: yamlast.ScalarNode, Value: ""},
		&fileNode)

	join := yamlast.Node{Kind: yamlast.MappingNode}
	join.Children = append(join.Children,
		&yamlast.Node{Kind: yamlast.ScalarNode, Value: "Fn::Join"},
		&joinArgs)

	return &join, nil
}

func vaultHandler(config *Config, tag string, value string) (*yamlast.Node, error) {

	if config.VaultAST == nil {
		return &yamlast.Node{Kind: yamlast.ScalarNode, Value: ""}, nil
	}

	node := yamlast.SelectNode(config.VaultAST, value)

	if node == nil {
		node = &yamlast.Node{Kind: yamlast.ScalarNode, Value: ""}
	}
	return node, nil
}

// Converts a template to a json string.
func templateToJSON(node *yamlast.Node) string {
	jsonData, err := json.MarshalIndent(nodeToInterface(node, nil), "", "  ")
	if err != nil {
		panic(err)
	}

	return string(jsonData)
}

// Converts a node to an object
func nodeToInterface(node *yamlast.Node, parent *yamlast.Node) interface{} {
	switch node.Kind {
	case yamlast.DocumentNode:
		if len(node.Children) > 0 {
			return nodeToInterface(node.Children[0], node)
		}
		return nil

	case yamlast.MappingNode:
		mapping := map[string]interface{}{}

		for i := 0; i < len(node.Children)/2; i++ {
			key := node.Children[i*2]
			value := node.Children[i*2+1]

			// Filter out metadata nodes
			if parent.Kind != yamlast.DocumentNode || key.Value != "CFToolMetadata" {
				mapping[key.Value] = nodeToInterface(value, node)
			}
		}
		return mapping

	case yamlast.SequenceNode:
		sequence := []interface{}{}

		for _, child := range node.Children {
			sequence = append(sequence, nodeToInterface(child, node))
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
