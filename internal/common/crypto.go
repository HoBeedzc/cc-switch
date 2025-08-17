package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	PBKDF2Iterations = 100000
	SaltLength       = 16
	KeyLength        = 32
	NonceLength      = 12
)

// EncryptionData holds encrypted data with salt and nonce
type EncryptionData struct {
	Salt      []byte
	Nonce     []byte
	Encrypted []byte
}

// DeriveKey derives encryption key from password using PBKDF2
func DeriveKey(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, PBKDF2Iterations, KeyLength, sha256.New)
}

// GenerateSalt generates a random salt for PBKDF2
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltLength)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

// GenerateNonce generates a random nonce for AES-GCM
func GenerateNonce() ([]byte, error) {
	nonce := make([]byte, NonceLength)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	return nonce, nil
}

// EncryptData encrypts data using AES-256-GCM with password
func EncryptData(data []byte, password string) (*EncryptionData, error) {
	salt, err := GenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	key := DeriveKey(password, salt)
	defer clearBytes(key) // Clear key from memory

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce, err := GenerateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	encrypted := gcm.Seal(nil, nonce, data, nil)

	return &EncryptionData{
		Salt:      salt,
		Nonce:     nonce,
		Encrypted: encrypted,
	}, nil
}

// DecryptData decrypts data using AES-256-GCM with password
func DecryptData(encData *EncryptionData, password string) ([]byte, error) {
	key := DeriveKey(password, encData.Salt)
	defer clearBytes(key) // Clear key from memory

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	decrypted, err := gcm.Open(nil, encData.Nonce, encData.Encrypted, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (wrong password?): %w", err)
	}

	return decrypted, nil
}

// clearBytes securely clears a byte slice from memory
func clearBytes(data []byte) {
	for i := range data {
		data[i] = 0
	}
}
