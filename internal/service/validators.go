package service

import (
	"fmt"
	"net/url"
)

func validateURL(urlString string) error {
	u, err := url.Parse(urlString)
	if err != nil {
		return fmt.Errorf("invalid feed URL: %w: %v", ErrInvalidURL, err)
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("%s: invalid URL scheme or host", ErrInvalidURL)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%s: URL scheme must be http or https", ErrInvalidURL)
	}
	return nil
}
