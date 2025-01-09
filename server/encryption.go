package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

func canCrypt() bool {
	if GetConfig().CryptKey == "" || len(GetConfig().CryptKey) != 32 {
		return false
	}
	return true
}

func encryptString(s string) string {
	aes, err := aes.NewCipher([]byte(GetConfig().CryptKey))
	if err != nil {
		panic(err)
	}
	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		panic(err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err)
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(s), nil)
	res := base64.StdEncoding.EncodeToString(ciphertext)
	return res
}

func decryptString(s string) string {
	ciphertext, _ := base64.StdEncoding.Strict().DecodeString(s)
	aes, err := aes.NewCipher([]byte(GetConfig().CryptKey))
	if err != nil {
		panic(err)
	}
	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		panic(err)
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, []byte(nonce), []byte(ciphertext), nil)
	if err != nil {
		panic(err)
	}

	return string(plaintext)
}
