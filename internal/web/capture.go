package web

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
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

func capturePageToBytes(input string) ([]byte, error) {
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

	resp, err := client.Get(input)
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
	body, err := capturePageToBytes(input)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
func capturePageToUnmarshal(url string, data any, unmarshal func([]byte, any) error) error {
	page, err := capturePageToBytes(url)
	if err != nil {
		return err
	}
	return unmarshal(page, data)
}

func CapturePageToJson(url string, data any) error {
	return capturePageToUnmarshal(url, data, json.Unmarshal)
}

func CapturePageToXml(url string, data any) error {
	return capturePageToUnmarshal(url, data, xml.Unmarshal)
}
