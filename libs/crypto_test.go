package libs

import (
	"context"
	"testing"
)

// TestEncryptDecrypt is a test function for the Encrypt and Decrypt functions.
func TestEncryptDecrypt(t *testing.T) {
	// Define the plain text to be encrypted.
	plainText := "Hello, world!"

	// Define the encryption key.
	key := "6e78049c36e8c38b7f72d8c770ce810264f1fc447120aba7519555f658c2fc67"

	// Create a new background context.
	ctx := context.Background()

	// Encrypt the plain text using the Encrypt function.
	encryptedString, err := Encrypt(ctx, plainText, key)
	if err != nil {
		t.Errorf("Encrypt failed with error: %v", err)
	}

	// Decrypt the encrypted string using the Decrypt function.
	decryptedString, err := Decrypt(ctx, encryptedString, key)
	if err != nil {
		t.Errorf("Decrypt failed with error: %v", err)
	}

	// Check if the decrypted string matches the original plain text.
	if decryptedString != plainText {
		t.Errorf("Decrypted string does not match the original plain text")
	}
}
