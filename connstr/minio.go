package connstr

import "net/url"

type MinioOptions struct {
	Secure    bool
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
}

func (o MinioOptions) String() string {
	var scheme string
	if o.Secure {
		scheme = "https"
	} else {
		scheme = "http"
	}
	u := url.URL{
		Scheme: scheme,
		Host:   o.Endpoint,
		User:   url.UserPassword(o.AccessKey, o.SecretKey),
	}
	u.Path = o.Bucket
	return u.String()
}
