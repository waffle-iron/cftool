package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

type commandHandler func(*Config)

func main() {
	commands := map[string]commandHandler{
		"process": processCmd,
		"vault":   vaultCmd,
	}

	flag.Parse()

	config := LoadConfig()

	command := flag.Arg(0)
	handler, ok := commands[command]
	if ok {
		handler(config)
	} else {
		usage(commands)
	}
}

func processCmd(config *Config) {
	templatePath := flag.Arg(1)
	template := NewTemplate(config)
	err := template.LoadFile(templatePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "An error occurred while processing the template: ", err.Error())
		os.Exit(-1)
	}

	fmt.Println(template.ToJSON())
}

func vaultCmd(config *Config) {
	command := flag.Arg(1)
	if command == "keygen" {
		vaultKeygenCommand(config)
	} else if command == "encrypt" {
		vaultEncryptCmd(config)
	} else if command == "decrypt" {
		vaultDecryptCmd(config)
	} else {
		fmt.Println("Vault Usage:")
		fmt.Println()
		fmt.Println("\tcftool vault [command]")
		fmt.Println()
		fmt.Println("Available Commands:")
		fmt.Println()
		fmt.Println("\tencrypt - Encrypt a vault file.")
		fmt.Println("\tdecrypt - Decrypt a vault file.")
		fmt.Println("\tkeygen - Generate a key.")
		fmt.Println()
	}
}

func vaultKeygenCommand(config *Config) {
	key, err := GenerateKey()

	if err != nil {
		fmt.Println("Error generating key:", err.Error())
		os.Exit(-1)
	}

	fmt.Println(EncodeVaultKey(key))
}

func vaultEncryptCmd(config *Config) {
	if config.VaultKey == nil {
		fmt.Println("Error loading vault key")
		os.Exit(-1)
	}

	source := flag.Arg(2)
	if source == "" {
		fmt.Println("Usage: cftool vault encrypt [encryptionSource]")
		fmt.Println()
		os.Exit(-1)
	}

	message, err := ioutil.ReadFile(source)
	if err != nil {
		fmt.Println("Error reading encryption source", source, ":", err.Error())
		fmt.Println()
		os.Exit(-1)
	}

	fmt.Println(Encrypt(string(message), config.VaultKey))
}

func vaultDecryptCmd(config *Config) {

	if config.VaultKey == nil {
		fmt.Println("Error opening .vaultkey: ")
		fmt.Println()
		os.Exit(-1)
	}

	source := flag.Arg(2)
	if source == "" {
		fmt.Println("Usage: cftool vault decrypt [encryptedFile]")
		fmt.Println()
		os.Exit(-1)
	}

	message, err := ioutil.ReadFile(source)
	if err != nil {
		fmt.Println("Error reading encrypted file", source, ":", err.Error())
		fmt.Println()
		os.Exit(-1)
	}

	fmt.Println(Decrypt(string(message), config.VaultKey))
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
