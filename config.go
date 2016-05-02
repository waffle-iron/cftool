package main

import (
	"io/ioutil"

	"github.com/commondream/yaml-ast"
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
		encryptedVaultData, vaultFileErr := ioutil.ReadFile("vault")
		if vaultFileErr == nil {
			decryptedVault := Decrypt(string(encryptedVaultData), config.VaultKey)
			config.VaultAST = yamlast.Parse([]byte(decryptedVault))
		}
	}

	return &config
}
