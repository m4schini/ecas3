package digest

type Digest string

func (d Digest) String() string {
	return string(d)
}
