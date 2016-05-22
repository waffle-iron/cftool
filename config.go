package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/commondream/yamlast"
)

// Config represents the configuration of cftool for an execution.
type Config struct {
	VaultKey []byte
	VaultAST *yamlast.Node
}

// LoadConfig loads a config.
func LoadConfig() *Config {
	config := Config{}

	vaultKey, keyErr := LoadVaultKey()
	if keyErr == nil {
		config.VaultKey = vaultKey
	}

	if config.VaultKey != nil {
		path := "vault"
		encryptedVaultData, vaultFileErr := ioutil.ReadFile(path)

		if vaultFileErr == nil && len(encryptedVaultData) > 0 {
			decryptedVault := Decrypt(string(encryptedVaultData), config.VaultKey)

			var err error
			config.VaultAST, err = yamlast.Parse([]byte(decryptedVault))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing vault yaml: %s\n", err.Error())
			}
		}
	}

	return &config
}
