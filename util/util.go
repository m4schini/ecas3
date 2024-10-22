package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// SizeCountingWriter is a custom writer that wraps another writer and counts the bytes written.
type SizeCountingWriter struct {
	Writer io.Writer
	Count  int64
}

// Write method to count the bytes and forward the write to the underlying writer.
func (scw *SizeCountingWriter) Write(p []byte) (n int, err error) {
	n, err = scw.Writer.Write(p)
	scw.Count += int64(n)
	return n, err
}

func GenerateKey(size int, encryptionKey []byte) (unencrypted, encrypted []byte, err error) {
	// Create a byte slice of the desired size
	unencrypted = make([]byte, size)
	// Read random bytes into the slice
	_, err = rand.Read(unencrypted)
	if err != nil {
		return nil, nil, err
	}

	encrypted, err = Encrypt(unencrypted, encryptionKey)
	if err != nil {
		return nil, nil, err
	}

	return unencrypted, encrypted, nil
}

// Encrypt a string using AES-256 GCM
func Encrypt(plaintext, key []byte) ([]byte, error) {
	// Create AES block cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Use GCM (Galois/Counter Mode)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Create a nonce of the appropriate size
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt the plaintext with the nonce and additional data (AAD is optional here)
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Encode to base64 for easier transmission
	return ciphertext, nil
}

// Decrypt a string using AES-256 GCM
func Decrypt(ciphertext, key []byte) ([]byte, error) {
	// Create AES block cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Use GCM (Galois/Counter Mode)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Extract the nonce from the ciphertext
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertextData := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertextData, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
