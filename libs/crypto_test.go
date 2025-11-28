package libs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func generateValidHexKey(t *testing.T) string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		t.Fatalf("failed to generate random key: %v", err)
	}
	return hex.EncodeToString(bytes)
}

func TestEncrypt_ValidInput(t *testing.T) {
	ctx := context.Background()
	key := generateValidHexKey(t)
	plaintext := "Hello, World!"

	result, err := Encrypt(ctx, plaintext, key)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.NotEqual(t, plaintext, result)
}

func TestEncrypt_InvalidHexKey(t *testing.T) {
	ctx := context.Background()
	invalidKey := "not-valid-hex"
	plaintext := "Hello, World!"

	result, err := Encrypt(ctx, plaintext, invalidKey)

	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.Contains(t, err.Error(), "decode encryption key from hex")
}

func TestEncrypt_WrongKeyLength(t *testing.T) {
	ctx := context.Background()
	shortKey := hex.EncodeToString([]byte("short"))
	plaintext := "Hello, World!"

	result, err := Encrypt(ctx, plaintext, shortKey)

	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.Contains(t, err.Error(), "failed to create cipher block")
}

func TestEncrypt_EmptyPlaintext(t *testing.T) {
	ctx := context.Background()
	key := generateValidHexKey(t)
	plaintext := ""

	result, err := Encrypt(ctx, plaintext, key)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestEncrypt_LongPlaintext(t *testing.T) {
	ctx := context.Background()
	key := generateValidHexKey(t)
	plaintext := "a"
	for i := 0; i < 10000; i++ {
		plaintext += "b"
	}

	result, err := Encrypt(ctx, plaintext, key)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestDecrypt_ValidInput(t *testing.T) {
	ctx := context.Background()
	key := generateValidHexKey(t)
	originalPlaintext := "Hello, World!"

	encrypted, err := Encrypt(ctx, originalPlaintext, key)
	assert.NoError(t, err)

	decrypted, err := Decrypt(ctx, encrypted, key)

	assert.NoError(t, err)
	assert.Equal(t, originalPlaintext, decrypted)
}

func TestDecrypt_InvalidHexKey(t *testing.T) {
	ctx := context.Background()
	invalidKey := "not-valid-hex"
	validKey := generateValidHexKey(t)
	encrypted, _ := Encrypt(ctx, "Hello", validKey)

	result, err := Decrypt(ctx, encrypted, invalidKey)

	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.Contains(t, err.Error(), "decode encryption key from hex")
}

func TestDecrypt_InvalidHexCiphertext(t *testing.T) {
	ctx := context.Background()
	key := generateValidHexKey(t)
	invalidCiphertext := "not-valid-hex"

	result, err := Decrypt(ctx, invalidCiphertext, key)

	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.Contains(t, err.Error(), "decode ciphertext from hex")
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	ctx := context.Background()
	key := generateValidHexKey(t)
	plaintext := "Hello, World!"

	encrypted, _ := Encrypt(ctx, plaintext, key)

	encBytes, _ := hex.DecodeString(encrypted)
	if len(encBytes) > 1 {
		encBytes[len(encBytes)-1] ^= 0xFF
	}
	tamperedEncrypted := hex.EncodeToString(encBytes)

	result, err := Decrypt(ctx, tamperedEncrypted, key)

	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.Contains(t, err.Error(), "decryption failed")
}

func TestDecrypt_WrongKey(t *testing.T) {
	ctx := context.Background()
	key1 := generateValidHexKey(t)
	key2 := generateValidHexKey(t)
	plaintext := "Hello, World!"

	encrypted, _ := Encrypt(ctx, plaintext, key1)

	result, err := Decrypt(ctx, encrypted, key2)

	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.Contains(t, err.Error(), "decryption failed")
}

func TestDecrypt_TooShortCiphertext(t *testing.T) {
	ctx := context.Background()
	key := generateValidHexKey(t)

	shortCiphertext := hex.EncodeToString([]byte("short"))

	result, err := Decrypt(ctx, shortCiphertext, key)

	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.Contains(t, err.Error(), "ciphertext too short")
}

func TestDecrypt_WrongKeyLength(t *testing.T) {
	ctx := context.Background()
	shortKey := hex.EncodeToString([]byte("short"))
	validKey := generateValidHexKey(t)
	encrypted, _ := Encrypt(ctx, "Hello", validKey)

	result, err := Decrypt(ctx, encrypted, shortKey)

	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.Contains(t, err.Error(), "failed to create cipher block")
}

func TestEncryptDecryptRoundTrip_VariousInputs(t *testing.T) {
	ctx := context.Background()
	key := generateValidHexKey(t)

	testCases := []struct {
		name      string
		plaintext string
	}{
		{"empty string", ""},
		{"single character", "a"},
		{"simple text", "Hello, World!"},
		{"special characters", "!@#$%^&*()_+-=[]{}|;:',.<>?/"},
		{"unicode", "你好世界🌍"},
		{"multiline", "line1\nline2\nline3"},
		{"long text", "Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
			"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encrypted, err := Encrypt(ctx, tc.plaintext, key)
			assert.NoError(t, err, "Encrypt should not error for: %s", tc.name)

			decrypted, err := Decrypt(ctx, encrypted, key)
			assert.NoError(t, err, "Decrypt should not error for: %s", tc.name)
			assert.Equal(t, tc.plaintext, decrypted, "Decrypted text should match original for: %s", tc.name)
		})
	}
}

func TestEncrypt_NoDeterministicOutput(t *testing.T) {
	ctx := context.Background()
	key := generateValidHexKey(t)
	plaintext := "Same text"

	encrypted1, err1 := Encrypt(ctx, plaintext, key)
	assert.NoError(t, err1)

	encrypted2, err2 := Encrypt(ctx, plaintext, key)
	assert.NoError(t, err2)

	assert.NotEqual(t, encrypted1, encrypted2, "Encryption should not be deterministic (nonce should vary)")

	decrypted1, _ := Decrypt(ctx, encrypted1, key)
	decrypted2, _ := Decrypt(ctx, encrypted2, key)
	assert.Equal(t, decrypted1, plaintext)
	assert.Equal(t, decrypted2, plaintext)
}

func TestDecrypt_EmptyString(t *testing.T) {
	ctx := context.Background()
	key := generateValidHexKey(t)

	result, err := Decrypt(ctx, "", key)

	assert.Error(t, err)
	assert.Equal(t, "", result)
}
