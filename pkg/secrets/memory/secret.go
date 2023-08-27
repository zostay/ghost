package memory

import (
	"net/url"
	"strconv"
	"time"

	"github.com/zostay/ghost/pkg/secrets"
)

func SecretMap(in secrets.Secret) map[string]string {
	out := make(map[string]string, len(in.Fields())+8)
	for k, v := range in.Fields() {
		out[k] = v
	}

	out["ID"] = in.ID()
	out["Name"] = in.Name()
	out["Username"] = in.Username()
	out["Password"] = in.Password()
	out["Type"] = in.Type()
	out["LastModified"] = strconv.FormatInt(in.LastModified().Unix(), 10)
	out["Location"] = in.Location()
	out["Url"] = secrets.UrlString(in)

	return out
}

func MapSecret(in map[string]string) secrets.Secret {
	var lm time.Time
	lmInt, err := strconv.ParseInt(in["LastModified"], 10, 64)
	if err != nil {
		lm = time.Unix(lmInt, 0)
	}

	var u *url.URL
	if us, ok := in["Url"]; ok {
		u, _ = url.Parse(us)
	}

	opts := make([]secrets.SingleOption, 0, len(in))
	for k, v := range in {
		switch k {
		case "ID":
			opts = append(opts, secrets.WithID(v))
		case "Name":
			opts = append(opts, secrets.WithName(v))
		case "Username":
			opts = append(opts, secrets.WithUsername(v))
		case "Password":
			opts = append(opts, secrets.WithPassword(v))
		case "Type":
			opts = append(opts, secrets.WithType(v))
		case "LastModified":
			opts = append(opts, secrets.WithLastModified(lm))
		case "Location":
			opts = append(opts, secrets.WithLocation(v))
		case "Url":
			opts = append(opts, secrets.WithUrl(u))
		default:
			opts = append(opts, secrets.WithField(k, v))
		}
	}

	return secrets.NewSecret("", "", "", opts...)
}
