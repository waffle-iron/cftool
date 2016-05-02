package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/commondream/yaml-ast"
)

type tagHandler func(*Config, string, string) *yamlast.Node

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

func loadTemplate(path string, config *Config) *yamlast.Node {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file %s.\n", path)
		os.Exit(1)
	}

	doc := yamlast.Parse(b)
	processTags(doc, config)
	return doc
}

func processTags(node *yamlast.Node, config *Config) {
	for index, child := range node.Children {
		if child.Tag != "" {
			handler := getTagHandler(child.Tag)
			if handler != nil {
				node.Children[index] = handler(config, child.Tag, child.Value)
			}
		}

		processTags(child, config)
	}
}

func importTagHandler(config *Config, tag string, value string) *yamlast.Node {
	subDoc := loadTemplate(fmt.Sprintf("./imports/%s.yml", value), config)
	return subDoc.Children[0]
}

func refHandler(config *Config, tag string, value string) *yamlast.Node {
	refNode := yamlast.Node{Kind: yamlast.MappingNode}
	refNode.Children = append(refNode.Children,
		&yamlast.Node{Kind: yamlast.ScalarNode, Value: "Ref"})
	refNode.Children = append(refNode.Children,
		&yamlast.Node{Kind: yamlast.ScalarNode, Value: value})

	return &refNode
}

func fileHandler(config *Config, tag string, value string) *yamlast.Node {
	path := fmt.Sprintf("./files/%s", value)
	file, err := os.Open(path)
	if err != nil {
		panic(fmt.Sprintf("Error loading file: %s", path))
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if scanner.Err() != nil {
		panic(fmt.Sprintf("Error loading file %s", path))
	}

	fileNode := yamlast.Node{Kind: yamlast.SequenceNode}
	for _, line := range lines {
		fileNode.Children = append(fileNode.Children,
			&yamlast.Node{Kind: yamlast.ScalarNode, Value: line})
	}

	return &fileNode
}

func vaultHandler(config *Config, tag string, value string) *yamlast.Node {

	if config.VaultAST == nil {
		return &yamlast.Node{Kind: yamlast.ScalarNode, Value: ""}
	}

	node := yamlast.SelectNode(config.VaultAST, value)

	if node == nil {
		node = &yamlast.Node{Kind: yamlast.ScalarNode, Value: ""}
	}
	return node
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
