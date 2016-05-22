package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/commondream/yamlast"
)

const (
	MetadataKey = "CFToolMetadata"
)

type tagHandler func(string, string) (*yamlast.Node, error)

func (template *Template) getTagHandler(tag string) tagHandler {
	switch tag {
	case "!import":
		return template.importTagHandler
	case "!ref":
		return template.refHandler
	case "!file":
		return template.fileHandler
	case "!vault":
		return template.vaultHandler
	case "!meta":
		return template.metadataHandler
	default:
		return nil
	}
}

// Template represents a template that we're proceesing.
type Template struct {
	Config       *Config
	DocumentNode *yamlast.Node
}

// NewTemplate initializes and returns a new template.
func NewTemplate(config *Config) *Template {
	template := &Template{Config: config}

	return template
}

func (template *Template) LoadFile(path string) error {
	_, err := template.loadFileInternal(path, true)
	return err
}

func (template *Template) loadFileInternal(path string, isRoot bool) (*yamlast.Node, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error reading file %s: %s", path, err))
	}

	return template.loadSourceInternal(b, isRoot)
}

func (template *Template) LoadSource(source []byte) error {
	_, err := template.loadSourceInternal(source, true)
	return err
}

func (template *Template) loadSourceInternal(source []byte, isRoot bool) (*yamlast.Node, error) {
	doc, err := yamlast.Parse(source)
	if err != nil {
		return nil, err
	}

	if isRoot {
		template.DocumentNode = doc
	}

	err = template.processTree(doc)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func (template *Template) processTree(node *yamlast.Node) error {
	for index, child := range node.Children {
		if child.Tag != "" {
			handler := template.getTagHandler(child.Tag)
			if handler != nil {
				var err error
				node.Children[index], err = handler(child.Tag, child.Value)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("Unknown tag: %s", child.Tag)
			}
		}

		err := template.processTree(child)
		if err != nil {
			return err
		}
	}

	return nil
}

func (template *Template) importTagHandler(tag string, value string) (*yamlast.Node, error) {
	subDoc, err := template.loadFileInternal(fmt.Sprintf("./imports/%s.yml", value), false)
	if err != nil {
		return nil, err
	}

	return subDoc.Children[0], nil
}

func (template *Template) refHandler(tag string, value string) (*yamlast.Node, error) {
	refNode := yamlast.Node{Kind: yamlast.MappingNode}
	refNode.Children = append(refNode.Children,
		&yamlast.Node{Kind: yamlast.ScalarNode, Value: "Ref"})
	refNode.Children = append(refNode.Children,
		&yamlast.Node{Kind: yamlast.ScalarNode, Value: value})

	return &refNode, nil
}

func (template *Template) fileHandler(tag string, value string) (*yamlast.Node, error) {
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

func (template *Template) vaultHandler(tag string, value string) (*yamlast.Node, error) {
	if template.Config.VaultAST == nil {
		return &yamlast.Node{Kind: yamlast.ScalarNode, Value: ""}, nil
	}

	node := yamlast.SelectNode(template.Config.VaultAST, value)

	if node == nil {
		node = &yamlast.Node{Kind: yamlast.ScalarNode, Value: ""}
	}
	return node, nil
}

func (template *Template) metadataNode() *yamlast.Node {
	topMap := template.DocumentNode.Children[0]

	if topMap.Kind != yamlast.MappingNode {
		return nil
	}

	for i := 0; i < len(topMap.Children)/2; i++ {
		if topMap.Children[i].Value == MetadataKey {
			return topMap.Children[i+1]
		}
	}

	return nil
}

func (template *Template) metadataHandler(tag string, value string) (*yamlast.Node, error) {
	metadataNode := template.metadataNode()
	if metadataNode != nil {
		node := yamlast.SelectNode(metadataNode, value)
		if node != nil {
			return node, nil
		}
	}

	return nil, fmt.Errorf("Unknown metadata value: %s", value)
}

// Converts a template to a json string.
func (template *Template) ToJSON() string {
	jsonData, err := json.MarshalIndent(nodeToInterface(template.DocumentNode, nil), "", "  ")
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
			if parent.Kind != yamlast.DocumentNode || key.Value != MetadataKey {
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
