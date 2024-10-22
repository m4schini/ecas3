package ecas

import (
	"context"
	"encoding/base64"
	"github.com/m4schini/ecas3/connstr"
	"github.com/m4schini/ecas3/crypto"
	"github.com/m4schini/ecas3/digest"
	"github.com/m4schini/ecas3/minio"
	"io"
	"os"
)

const (
	ConnStrEnvName = "CAS_CONNECTION_STRING"
	SecretEnvName  = "CAS_STATIC_SECRET"
)

type Store interface {
	Put(ctx context.Context, contentType string, reader io.Reader) (digest digest.Digest, err error)
	Get(ctx context.Context, digest digest.Digest) (reader io.ReadCloser, size int64, contentType string, err error)
	Exists(ctx context.Context, digest digest.Digest) (exists bool, err error)
	Remove(ctx context.Context, digest digest.Digest) (err error)
	io.Closer
}

func FromConnString(connString string, secret []byte) (Store, error) {
	options, err := connstr.Parse(connString)
	if err != nil {
		return nil, err
	}
	client, err := minio.NewClient(options.Secure, options.Endpoint, options.AccessKey, options.SecretKey)
	if err != nil {
		return nil, err
	}
	store := &minio.Store{
		Client:     client,
		BucketName: options.Bucket,
		Secret:     &crypto.StaticSecret{Secret: secret},
	}
	return store, nil
}

func FromEnv() (Store, error) {
	conn := os.Getenv(ConnStrEnvName)
	secret, err := base64.StdEncoding.DecodeString(os.Getenv(SecretEnvName))
	if err != nil {
		return nil, err
	}
	return FromConnString(conn, secret)
}
