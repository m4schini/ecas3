package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

// AESCFBEncryptionWriter is a custom io.Writer that encrypts data before writing.
type AESCFBEncryptionWriter struct {
	writer io.Writer
	stream cipher.Stream
}

// NewWriter creates a new AESCFBEncryptionWriter.
func NewWriter(w io.Writer, key []byte) (*AESCFBEncryptionWriter, error) {
	// AES block cipher requires keys of 16, 24, or 32 bytes
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create an initialization vector (IV)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// Write the IV to the underlying writer, it will be needed for decryption
	if _, err := w.Write(iv); err != nil {
		return nil, err
	}

	// Create a CFB stream cipher for encryption
	stream := cipher.NewCFBEncrypter(block, iv)

	return &AESCFBEncryptionWriter{
		writer: w,
		stream: stream,
	}, nil
}

// Write encrypts data and writes it to the underlying writer.
func (ew *AESCFBEncryptionWriter) Write(p []byte) (n int, err error) {
	encrypted := make([]byte, len(p))
	ew.stream.XORKeyStream(encrypted, p)
	return ew.writer.Write(encrypted)
}
