package web

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/sync/semaphore"
)

var (
	client     = retryablehttp.NewClient()
	permits    = semaphore.NewWeighted(150)
	urlPermits = map[string]*semaphore.Weighted{
		"sourceforge.net": semaphore.NewWeighted(5),
	}
)

func capturePageToBytes(input string, headers http.Header) ([]byte, error) {
	url, err := url.Parse(input)
	if err != nil {
		return nil, err
	}
	if sem, exists := urlPermits[url.Hostname()]; exists {
		if err := sem.Acquire(context.Background(), 1); err != nil {
			return nil, err
		}
		defer sem.Release(1)
	}
	if err := permits.Acquire(context.Background(), 1); err != nil {
		return nil, err
	}
	defer permits.Release(1)

	req, err := retryablehttp.NewRequest("GET", input, nil)
	if err != nil {
		return nil, err
	}
	req.Header = headers
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Failed to capture page %s: %s", input, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func CapturePage(input string) (string, error) {
	body, err := capturePageToBytes(input, nil)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
func capturePageToUnmarshal(url string, data any, unmarshal func([]byte, any) error, headers http.Header) error {
	page, err := capturePageToBytes(url, headers)
	if err != nil {
		return err
	}
	return unmarshal(page, data)
}

func CapturePageToJson(url string, data any) error {
	return capturePageToUnmarshal(url, data, json.Unmarshal, nil)
}

func CapturePageAcceptingJson(url string, data any) error {
	return capturePageToUnmarshal(url, data, json.Unmarshal, http.Header{"Accept": []string{"application/json"}})
}

func CapturePageToXml(url string, data any) error {
	return capturePageToUnmarshal(url, data, xml.Unmarshal, nil)
}

func FinalRedirectUrl(input string) (string, error) {
	url, err := url.Parse(input)
	if err != nil {
		return "", err
	}

	if sem, exists := urlPermits[url.Hostname()]; exists {
		if err := sem.Acquire(context.Background(), 1); err != nil {
			return "", err
		}
		defer sem.Release(1)
	}
	if err := permits.Acquire(context.Background(), 1); err != nil {
		return "", err
	}
	defer permits.Release(1)

	resp, err := client.Get(input)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("Failed to find Redirect URL for page %s: %s", input, resp.Status)
	}

	return resp.Request.URL.String(), nil
}
