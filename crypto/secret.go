package crypto

type SecretStorage interface {
	Get() ([]byte, error)
}

type StaticSecret struct{ Secret []byte }

func (s *StaticSecret) Get() ([]byte, error) {
	return s.Secret, nil
}
