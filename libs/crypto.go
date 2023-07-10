package libs

// courtesy of https://www.melvinvivas.com/how-to-encrypt-and-decrypt-data-using-aes

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"

	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slog"
)

var tracerCrypto = otel.Tracer("crypto")

// Encrypt encrypts a string using AES-GCM algorithm with a given key. The key
// should be provided as a hexadecimal string. It returns the encrypted string
// in hexadecimal format.
func Encrypt(ctx context.Context, stringToEncrypt string, keyString string) (string, error) {
	_, span := tracerCrypto.Start(ctx, "Encrypt")
	defer span.End()

	//Since the key is in string, we need to convert decode it to bytes
	key, _ := hex.DecodeString(keyString)
	plaintext := []byte(stringToEncrypt)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		slog.Error(err.Error())
		return "", err
	}

	//Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	//https://golang.org/pkg/crypto/cipher/#NewGCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		slog.Error(err.Error())
		return "", err
	}

	//Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		slog.Error(err.Error())
		return "", err
	}

	//Encrypt the data using aesGCM.Seal
	//Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return fmt.Sprintf("%x", ciphertext), nil
}

// Decrypt decrypts a previously encrypted string using the same key used to
// encrypt it. It takes in an encrypted string and a key string as parameters
// and returns the decrypted string. The key must be in hexadecimal format.
func Decrypt(ctx context.Context, encryptedString string, keyString string) (string, error) {
	_, span := tracerCrypto.Start(ctx, "Decrypt")
	defer span.End()

	key, _ := hex.DecodeString(keyString)
	enc, _ := hex.DecodeString(encryptedString)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		slog.Error(err.Error())
		return "", err
	}

	//Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		slog.Error(err.Error())
		return "", err
	}

	//Get the nonce size
	nonceSize := aesGCM.NonceSize()

	//Extract the nonce from the encrypted data
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	//Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		slog.Error(err.Error())
		return "", err
	}

	return string(plaintext), nil
}
