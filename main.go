package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/commondream/yaml-ast"
)

type commandHandler func()

func main() {
	commands := map[string]commandHandler{
		"process": processCmd,
	}

	flag.Parse()

	command := flag.Arg(0)
	handler, ok := commands[command]
	if ok {
		handler()
	} else {
		usage(commands)
	}
}

func processCmd() {
	template := flag.Arg(1)
	b, err := ioutil.ReadFile(template)
	if err != nil {
		fmt.Printf("Error reading file %s.\n", template)
		os.Exit(1)
	}

	doc := yamlast.Parse(b)
	jsonData, err := json.MarshalIndent(nodeToInterface(doc), "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Print(string(jsonData))
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

// Prints generic usage for the entire app
func usage(commands map[string]commandHandler) {
	fmt.Println("cftool - A helpful CloudFormation wrapper")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("\tcftool command [arguments]")
	fmt.Println()
	fmt.Println("Available commands:")
	for command := range commands {
		fmt.Printf("\t%s\n", command)
	}
	fmt.Println()
}
