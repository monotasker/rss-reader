package feed

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Fetch downloads the raw feed document.
// If successful, it returns the response as bytes, usually
// in XML format. For non-2xx responses it returns an error.
func Fetch(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Fetch %s: unexpected status %s", url, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Read body %s: %w", url, err)
	}

	return body, nil

}
