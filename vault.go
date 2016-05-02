package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io/ioutil"
	"strings"
)

const nonceLength = 12

// GenerateKey generates a random key and base64 encodes it
func GenerateKey() ([]byte, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)

	if err != nil {
		return []byte{}, err
	}

	return b, nil
}

// EncodeVaultKey Base64 encodes a vault key.
func EncodeVaultKey(key []byte) string {
	return base64.StdEncoding.EncodeToString(key)
}

// LoadVaultKey loads and returns a vault key.
func LoadVaultKey() ([]byte, error) {
	keyBase64, err := ioutil.ReadFile(".vaultkey")
	if err != nil {
		return []byte{}, err
	}

	key := make([]byte, base64.StdEncoding.DecodedLen(len(keyBase64)))
	_, err = base64.StdEncoding.Decode(key, keyBase64)
	if err != nil {
		return []byte{}, err
	}

	return key[:32], nil
}

// Encrypt takes a message and encrypts it with the vault key. Returns a
// base64 encoded encrypted message.
func Encrypt(message string, key []byte) string {
	nonce := make([]byte, nonceLength)
	_, err := rand.Read(nonce)
	if err != nil {
		panic(err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	ciphertext := aesgcm.Seal(nil, nonce, []byte(message), nil)
	out := append(nonce, ciphertext...)

	return base64.StdEncoding.EncodeToString(out)
}

// Decrypt takes in an encrypted base64 encoded string and key and returns the decrypted
func Decrypt(encryptedBase64 string, key []byte) string {
	encryptedBase64Bytes := []byte(strings.TrimSpace(encryptedBase64))

	encrypted := make([]byte, base64.StdEncoding.DecodedLen(len(encryptedBase64Bytes)))
	l, err := base64.StdEncoding.Decode(encrypted, encryptedBase64Bytes)
	if err != nil {
		panic(err)
	}
	encrypted = encrypted[:l]

	nonce := encrypted[:nonceLength]
	message := encrypted[nonceLength:]

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	decrypted, err := aesgcm.Open(nil, nonce, message, nil)
	if err != nil {
		panic(err.Error())
	}

	return string(decrypted)
}
