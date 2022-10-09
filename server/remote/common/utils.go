package common

import (
	"net"
	"net/url"
	"strings"
)

func ExtractHostFromCloneURL(cloneURL string) (string, error) {
	u, err := url.Parse(cloneURL)
	if err != nil {
		return "", err
	}

	if !strings.Contains(u.Host, ":") {
		return u.Host, nil
	}

	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		return "", err
	}

	return host, nil
}

// Paginate iterates over a func call until it does not return new items and return it as list
func Paginate[T any](get func(page int) ([]T, error)) ([]T, error) {
	items := make([]T, 0, 10)
	page := 1
	lenFirstBatch := -1

	for {
		batch, err := get(page)
		if err != nil {
			return nil, err
		}
		items = append(items, batch...)

		if page == 1 {
			lenFirstBatch = len(batch)
		} else if len(batch) < lenFirstBatch || len(batch) == 0 {
			break
		}

		page++
	}

	return items, nil
}
