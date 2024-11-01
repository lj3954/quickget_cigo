package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	quickgetdata "github.com/quickemu-project/quickget_configs/pkg/quickget_data"
	"golang.org/x/sync/semaphore"
)

type Config quickgetdata.Config

type OSData struct {
	Name        string
	PrettyName  string
	Homepage    string
	Description string
	Releases    []Config
}

type Distro interface {
	Data() OSData
	CreateConfigs(chan Failure) ([]Config, error)
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

	resp, err := client.Get(input)
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

func SingleWhitespaceChecksum(url string) (string, error) {
	data, err := CapturePage(url)
	if err != nil {
		return "", fmt.Errorf("Failed to find single checksum: %w", err)
	}
	return BuildSingleWhitespaceChecksum(data)
}

func BuildSingleWhitespaceChecksum(data string) (string, error) {
	index := strings.Index(data, " ")
	if index == -1 {
		return "", errors.New("No whitespace was present in the checksum data")
	}
	return data[0:index], nil
}

type ChecksumSeparation interface {
	BuildWithData(string) map[string]string
}

func BuildChecksum(cs ChecksumSeparation, url string) (map[string]string, error) {
	data, err := CapturePage(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to build checksums: %w", err)
	}
	return cs.BuildWithData(data), nil
}

type Whitespace struct{}
type Md5Regex struct{}
type Sha256Regex struct{}
type CustomRegex struct {
	Regex      *regexp.Regexp
	KeyIndex   int
	ValueIndex int
}

func (Whitespace) BuildWithData(data string) map[string]string {
	m := make(map[string]string, 0)
	for _, line := range strings.Split(data, "\n") {
		slice := strings.SplitN(line, " ", 2)
		if len(slice) == 2 {
			hash := strings.TrimSpace(slice[0])
			file := strings.TrimSpace(slice[1])
			m[file] = hash
		}
	}
	return m
}

var md5Regex = regexp.MustCompile(`MD5 \(([^)]+)\) = ([0-9a-f]+)`)

func (Md5Regex) BuildWithData(data string) map[string]string {
	return CustomRegex{
		Regex:      md5Regex,
		KeyIndex:   1,
		ValueIndex: 2,
	}.BuildWithData(data)
}

var sha256Regex = regexp.MustCompile(`SHA256 \(([^)]+)\) = ([0-9a-f]+)`)

func (Sha256Regex) BuildWithData(data string) map[string]string {
	return CustomRegex{
		Regex:      sha256Regex,
		KeyIndex:   1,
		ValueIndex: 2,
	}.BuildWithData(data)
}

func (re CustomRegex) BuildWithData(data string) map[string]string {
	m := make(map[string]string, 0)
	for _, match := range re.Regex.FindAllStringSubmatch(data, -1) {
		file := match[re.KeyIndex]
		hash := match[re.ValueIndex]
		m[file] = hash
	}
	return m
}

func GetChannels() (chan Config, sync.WaitGroup) {
	return make(chan Config), sync.WaitGroup{}
}

func WaitForConfigs(ch chan Config, wg *sync.WaitGroup) []Config {
	go func() {
		wg.Wait()
		close(ch)
	}()

	configs := make([]Config, 0)
	for config := range ch {
		configs = append(configs, config)
	}
	return configs
}

type GithubAPI struct {
	TagName    string        `json:"tag_name"`
	Assets     []GithubAsset `json:"assets"`
	Prerelease bool          `json:"prerelease"`
	Body       string        `json:"body"`
}

type GithubAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

type Failure struct {
	Release  string
	Edition  string
	Arch     quickgetdata.Arch
	Error    error
	Checksum bool
}
