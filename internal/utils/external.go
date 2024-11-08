package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"
	"sync"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	quickgetdata "github.com/quickemu-project/quickget_configs/pkg/quickget_data"
	"golang.org/x/sync/semaphore"
)

type Config quickgetdata.Config

type OSData struct {
	Name        string   `json:"name"`
	PrettyName  string   `json:"pretty_name"`
	Homepage    string   `json:"homepage"`
	Description string   `json:"description"`
	Releases    []Config `json:"releases"`
}

type Distro interface {
	Data() OSData
	CreateConfigs(chan Failure, chan Failure) ([]Config, error)
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
	client     = retryablehttp.NewClient()
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
	return data[:index], nil
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
type CustomRegex struct {
	Regex      *regexp.Regexp
	KeyIndex   int
	ValueIndex int
}

var Md5Regex = CustomRegex{
	Regex:      regexp.MustCompile(`MD5 \(([^)]+)\) = ([0-9a-f]+)`),
	KeyIndex:   1,
	ValueIndex: 2,
}
var Sha256Regex = CustomRegex{
	Regex:      regexp.MustCompile(`SHA256 \(([^)]+)\) = ([0-9a-f]+)`),
	KeyIndex:   1,
	ValueIndex: 2,
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
	Release string
	Edition string
	Arch    quickgetdata.Arch
	Error   error
}

func GetBasicReleases(url, pattern string, num int) ([]string, error) {
	page, err := CapturePage(url)
	if err != nil {
		return nil, err
	}
	releaseRe := regexp.MustCompile(pattern)
	matches := releaseRe.FindAllStringSubmatch(page, num)

	releases := make([]string, len(matches))
	for i, match := range matches {
		releases[i] = match[1]
	}

	return releases, nil
}
