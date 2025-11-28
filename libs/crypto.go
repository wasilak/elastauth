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
)

var tracerCrypto = otel.Tracer("crypto")

// Encrypt encrypts a plaintext string using the AES-256-GCM cipher.
// The key parameter must be a 64-character hexadecimal string (representing 32 bytes).
// Returns the ciphertext as a hexadecimal string with the nonce prepended.
func Encrypt(ctx context.Context, stringToEncrypt string, keyString string) (string, error) {
	_, span := tracerCrypto.Start(ctx, "Encrypt")
	defer span.End()

	key, err := hex.DecodeString(keyString)
	if err != nil {
		return "", fmt.Errorf("failed to decode encryption key from hex: %w", err)
	}
	plaintext := []byte(stringToEncrypt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return fmt.Sprintf("%x", ciphertext), nil
}

// Decrypt decrypts an AES-256-GCM encrypted string back to plaintext.
// The encryptedString parameter must be a hexadecimal string produced by Encrypt.
// The key parameter must be the same 64-character hexadecimal key used during encryption.
func Decrypt(ctx context.Context, encryptedString string, keyString string) (string, error) {
	_, span := tracerCrypto.Start(ctx, "Decrypt")
	defer span.End()

	key, err := hex.DecodeString(keyString)
	if err != nil {
		return "", fmt.Errorf("failed to decode encryption key from hex: %w", err)
	}

	enc, err := hex.DecodeString(encryptedString)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext from hex: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(enc) < nonceSize+16 {
		return "", fmt.Errorf("ciphertext too short: %d bytes (minimum %d bytes required)", len(enc), nonceSize+16)
	}

	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed (ciphertext may be corrupted or tampered): %w", err)
	}

	return string(plaintext), nil
}
