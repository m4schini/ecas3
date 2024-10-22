package connstr

import (
	"fmt"
	"net/url"
)

func Parse(str string) (MinioOptions, error) {
	u, err := url.Parse(str)
	if err != nil {
		return MinioOptions{}, err
	}

	var secure bool
	switch u.Scheme {
	case "https":
		secure = true
	case "http":
		secure = false
	default:
		return MinioOptions{}, fmt.Errorf("unsupported scheme %s", u.Scheme)
	}

	if len(u.Path) > 1 && u.Path[0] == '/' {
		u.Path = u.Path[1:]
	}

	user := u.User
	password, _ := u.User.Password()
	return MinioOptions{
		Secure:    secure,
		Endpoint:  u.Host,
		AccessKey: user.Username(),
		SecretKey: password,
		Bucket:    u.Path,
	}, err
}
