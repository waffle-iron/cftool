package main

import (
	"flag"
	"fmt"
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
	doc := loadTemplate(template)
	fmt.Println(templateToJSON(doc))
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
