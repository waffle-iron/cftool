package main

import (
	"testing"

	"github.com/stvp/assert"
)

func TestEncryptAndDecrypt(t *testing.T) {
	message := "THIS IS A TEST"
	key, err := GenerateKey()
	assert.Nil(t, err)

	encrypted := Encrypt(message, key)
	decrypted := Decrypt(encrypted, key)

	assert.Equal(t, message, decrypted)
}
