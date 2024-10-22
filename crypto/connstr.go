package crypto

import "net/url"

type Options struct {
	Scheme string
	Static string
}

func (o Options) String() string {
	var u url.URL
	switch o.Scheme {
	case "static":
		u.Scheme = "static"
		u.Host = o.Static
		return u.String()
	default:
		panic("not implemented")
	}
}
