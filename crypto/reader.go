package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
)

type AESCFBDecryptionReader struct {
	stream cipher.Stream
	reader io.Reader
	close  func() error
}

func (dr *AESCFBDecryptionReader) Read(p []byte) (int, error) {
	n, err := dr.reader.Read(p)
	if n > 0 {
		dr.stream.XORKeyStream(p[:n], p[:n]) // Decrypt the data as it's read
	}
	return n, err
}

func (dr *AESCFBDecryptionReader) Close() error {
	if dr.close != nil {
		return dr.close()
	}
	return nil
}

// NewReader returns an io.ReadCloser that decrypts the input stream using AES-CFB in a streaming fashion
func NewReader(r io.ReadCloser, key []byte) (io.ReadCloser, error) {
	return NewReaderWithCloser(r, r, key)
}

func NewReaderWithCloser(r io.Reader, c io.Closer, key []byte) (io.ReadCloser, error) {
	// AES block size is 16 bytes
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(r, iv); err != nil {
		return nil, err
	}

	// Create AES block cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create CFB decrypter stream
	stream := cipher.NewCFBDecrypter(block, iv)

	// Return a ReadCloser that decrypts as it reads
	return &AESCFBDecryptionReader{
		stream: stream,
		reader: r,
		close:  c.Close,
	}, nil
}
