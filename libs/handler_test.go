package libs

import (
	"context"
	"testing"
)

var encryptedPassword = "MGJhN2NiOTUzOTU5NjQ1NmNjZWY1NzI4OGI1NWVhMDA3M2NlNDc1ODlmMmYxODI4ZGJiOTFiMmQ4YWY2YzhhNWI3MDFjMzIwYTk0ODNhNjE2YzA4NGI2ZWE2YzkxMjJkYTM="

// The TestDecryptingPassword function tests the decryption of a password using a given key.
func TestDecryptingPassword(t *testing.T) {
	// Create a new background context.
	ctx := context.Background()

	// Set the key to be used for decryption.
	key := "4aae8bec2b798229ffae60a84f9f6eaf5b7e912301ff0794fa173641a4f21adf"

	// Call the decryptingPassword function passing in the context, key, and encryptedPassword.
	decryptedPassword, err := decryptingPassword(ctx, key, encryptedPassword)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Define the expected decrypted password.
	expectedDecryptedPassword := "super-secret-password"

	// Check if the decrypted password matches the expected decrypted password.
	if decryptedPassword != expectedDecryptedPassword {
		t.Errorf("Incorrect decrypted password. Expected: %s, got: %s", expectedDecryptedPassword, decryptedPassword)
	}
}
