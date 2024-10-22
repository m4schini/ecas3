package minio

import (
	"bufio"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/m4schini/ecas3/crypto"
	"github.com/m4schini/ecas3/digest"
	"github.com/m4schini/ecas3/util"
	"github.com/minio/minio-go/v7"
	"io"
	"os"
)

const (
	defaultKeyLength = 32
)

type Store struct {
	Client     *minio.Client
	BucketName string
	Secret     crypto.SecretStorage
}

func (m *Store) Put(ctx context.Context, contentType string, reader io.Reader) (hash digest.Digest, err error) {
	key, err := m.getSecret()
	if err != nil {
		return "", err
	}
	encryptionKey, encryptedEncryptionKey, err := util.GenerateKey(defaultKeyLength, key)
	if err != nil {
		return hash, err
	}
	f, hash, size, err := prepareFile(reader, encryptionKey)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if exists, _ := m.Exists(ctx, hash); exists {
		return hash, nil
	}

	_, err = m.Client.PutObject(ctx, m.BucketName, hash.String(), f, size, minio.PutObjectOptions{
		UserMetadata: map[string]string{
			"Secret": base64.StdEncoding.EncodeToString(encryptedEncryptionKey),
		},
		ContentType: contentType,
	})
	return hash, err
}

func (m *Store) Get(ctx context.Context, digest digest.Digest) (reader io.ReadCloser, size int64, contentType string, err error) {
	ch := make(chan minio.ObjectInfo, 1)
	defer close(ch)
	errCh := make(chan error, 1)
	defer close(errCh)
	go func() {
		info, err := m.Client.StatObject(ctx, m.BucketName, digest.String(), minio.StatObjectOptions{})
		if err != nil {
			errCh <- err
			return
		}
		ch <- info
	}()

	reader, err = m.Client.GetObject(ctx, m.BucketName, digest.String(), minio.GetObjectOptions{})
	if err != nil {
		return nil, size, "", err
	}

	buffer := bufio.NewReaderSize(reader, 500_000)

	var info minio.ObjectInfo
	select {
	case info = <-ch:
		contentType = info.ContentType
		break
	case err = <-errCh:
		return nil, size, "", err
	}

	for s, s2 := range info.UserMetadata {
		fmt.Println(s, s2)
	}

	secret, exists := info.UserMetadata["Secret"]
	if !exists {
		return nil, size, "", fmt.Errorf("no secret found in object")
	}
	secretKey, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return nil, size, "", err
	}
	key, err := m.getSecret()
	if err != nil {
		return nil, size, "", err
	}
	key, err = util.Decrypt(secretKey, key)
	if err != nil {
		return nil, size, "", err
	}

	reader, err = crypto.NewReaderWithCloser(buffer, reader, key)
	return reader, info.Size, contentType, err
}

func (m *Store) Exists(ctx context.Context, digest digest.Digest) (exists bool, err error) {
	_, err = m.Client.StatObject(ctx, m.BucketName, digest.String(), minio.StatObjectOptions{})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (m *Store) Remove(ctx context.Context, digest digest.Digest) (err error) {
	return m.Client.RemoveObject(ctx, m.BucketName, digest.String(), minio.RemoveObjectOptions{})
}

func (m *Store) Close() error {
	return nil
}

func (m *Store) getSecret() ([]byte, error) {
	if m.Secret == nil {
		return []byte{}, fmt.Errorf("secret not initialized")
	}
	return m.Secret.Get()
}

func prepareFile(reader io.Reader, encryptionKey []byte) (io.ReadCloser, digest.Digest, int64, error) {
	f, err := os.CreateTemp("", "cas_")
	if err != nil {
		return nil, "", 0, err
	}
	defer f.Close()
	bf := &bufferedFile{Name: f.Name()}
	hasher := sha512.New() // Create a new sha512 hash.Hash

	counter := &util.SizeCountingWriter{Writer: f}
	var writer io.Writer = counter
	writer, err = crypto.NewWriter(writer, encryptionKey)
	if err != nil {
		return nil, "", 0, err
	}
	multiwriter := io.MultiWriter(hasher, writer)
	if _, err := io.Copy(multiwriter, reader); err != nil {
		return nil, "", 0, err
	}

	// Compute the hash
	sum := hasher.Sum(nil)
	// Convert hash to a string
	hash := hex.EncodeToString(sum)
	d := digest.Digest(hash)

	return bf, d, counter.Count, nil
}

type bufferedFile struct {
	Name string
	file *os.File
}

func (b *bufferedFile) Read(p []byte) (n int, err error) {
	if b.file == nil {
		f, err := os.Open(b.Name)
		if err != nil {
			return 0, err
		}
		b.file = f
	}

	return b.file.Read(p)
}

func (b *bufferedFile) Close() error {
	if b.file == nil {
		b.file.Close()
	}
	os.Remove(b.Name)
	return nil
}
