package libs

import (
	"context"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	// Set up test data
	plainText := "Hello, world!"
	key := "6e78049c36e8c38b7f72d8c770ce810264f1fc447120aba7519555f658c2fc67"

	// Test Encrypt function
	ctx := context.Background()
	encryptedString, err := Encrypt(ctx, plainText, key)
	if err != nil {
		t.Errorf("Encrypt failed with error: %v", err)
	}

	// Test Decrypt function
	decryptedString, err := Decrypt(ctx, encryptedString, key)
	if err != nil {
		t.Errorf("Decrypt failed with error: %v", err)
	}

	// Check if the decrypted string matches the original plain text
	if decryptedString != plainText {
		t.Errorf("Decrypted string does not match the original plain text")
	}
}
