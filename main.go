package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/nacl/secretbox"
)

type commandHandler func()

func main() {
	commands := map[string]commandHandler{
		"process": processCmd,
		"vault":   vaultCmd,
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

func vaultCmd() {
	command := flag.Arg(1)
	if command == "keygen" {
		vaultKeygenCommand()
	} else if command == "encrypt" {
		vaultEncryptCmd()
	} else if command == "decrypt" {
		vaultDecryptCmd()
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

func vaultKeygenCommand() {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error generating random number for key.")
		fmt.Printf("\tmessage: %s", err.Error())
		os.Exit(-1)
	}

	encoded := base64.StdEncoding.EncodeToString(b)
	fmt.Println(encoded)
}

func loadVaultKey() [32]byte {
	keyBase64, err := ioutil.ReadFile(".vaultkey")
	if err != nil {
		panic(err.Error())
	}

	key := make([]byte, base64.StdEncoding.DecodedLen(len(keyBase64)))
	l, err := base64.StdEncoding.Decode(key, keyBase64)
	if err != nil {
		panic(err.Error())
	}

	var byteKey [32]byte
	copy(byteKey[:], key[:l])
	return byteKey
}

func vaultEncryptCmd() {
	key := loadVaultKey()

	// source := flag.Arg(2)
	// if path == "" {
	// 	fmt.Println("Usage: cftool vault encrypt [encryptionSource]")
	// 	fmt.Println()
	// 	os.Exit(-1)
	// }

	source := []byte("THIS IS A TEST")

	var nonce [24]byte
	_, err := rand.Read(nonce[:])
	if err != nil {
		panic(err)
	}

	encrypted := nonce[:]
	result := secretbox.Seal(encrypted, source, &nonce, &key)

	fmt.Println(base64.StdEncoding.EncodeToString(result))
}

func vaultDecryptCmd() {
	key := loadVaultKey()

	encryptedBase64 := []byte("i+Yy9WqzO/YB8GvqmxAH4I7stFiL/HLqEIMVkhnN2dPVn0HZRZRcx6tku3XTArbZfep1clyQ")

	encrypted := make([]byte, base64.StdEncoding.DecodedLen(len(encryptedBase64)))
	_, err := base64.StdEncoding.Decode(encrypted, encryptedBase64)
	if err != nil {
		panic(err)
	}

	var nonce [24]byte
	copy(nonce[:], encrypted[:24])

	decrypted, ok := secretbox.Open(nil, encrypted[24:], &nonce, &key)

	if ok {
		fmt.Println(string(decrypted))
	} else {
		panic("decrypt failed")
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
