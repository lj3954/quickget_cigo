package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	quickgetdata "github.com/quickemu-project/quickget_configs/pkg/quickget_data"
	"golang.org/x/sync/semaphore"
)

type Config quickgetdata.Config

type OSData struct {
	Name        string
	PrettyName  string
	Homepage    string
	Description string
}

type Distro interface {
	Data() OSData
	CreateConfigs() ([]Config, error)
}

func CapturePage(input string) (string, error) {
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
	permits.Acquire(context.Background(), 1)
	defer permits.Release(1)

	resp, err := http.Get(input)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("Failed to capture page %s: %s", input, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

var (
	client     = &http.Client{}
	permits    = semaphore.NewWeighted(150)
	urlPermits = map[string]*semaphore.Weighted{
		"sourceforge.net": semaphore.NewWeighted(5),
	}
)
