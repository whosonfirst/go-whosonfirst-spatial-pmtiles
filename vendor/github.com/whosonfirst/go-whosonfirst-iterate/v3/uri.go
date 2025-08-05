package iterate

import (
	"fmt"
	"net/url"
)

// ScrubURI attempts to remove sensitive data (parameters, etc.) from 'uri' and return a new string (URI)
// which is safe to include in logging messages.
func ScrubURI(uri string) (string, error) {

	to_scrub := []string{
		"access_token",
	}

	u, err := url.Parse(uri)

	if err != nil {
		return "", fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	for _, k := range to_scrub {

		if q.Has(k) {
			q.Del(k)
			q.Set(k, "...")
		}
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}
