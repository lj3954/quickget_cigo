package utils

import (
	"cmp"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"iter"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/go-version"
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
	permits.Acquire(context.Background(), 1)
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

var (
	client     = retryablehttp.NewClient()
	permits    = semaphore.NewWeighted(150)
	urlPermits = map[string]*semaphore.Weighted{
		"sourceforge.net": semaphore.NewWeighted(5),
	}
)

func GetChannels() (chan Config, *sync.WaitGroup) {
	return make(chan Config), &sync.WaitGroup{}
}

func GetChannelsWith(num int) (chan Config, *sync.WaitGroup) {
	ch, wg := GetChannels()
	wg.Add(num)
	return ch, wg
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

func GetSortedReleasesFunc(url string, pattern any, num int, cmp func(a, b string) int) ([]string, error) {
	page, err := CapturePage(url)
	if err != nil {
		return nil, err
	}
	releaseRe, err := toRegexp(pattern)
	if err != nil {
		return nil, err
	}
	matches := releaseRe.FindAllStringSubmatch(page, -1)
	releases := make([]string, len(matches))
	for i, match := range matches {
		releases[i] = match[1]
	}

	slices.SortFunc(releases, cmp)
	if num >= 0 {
		numReturns := min(len(releases), num)
		firstIndex := len(releases) - numReturns
		clear(releases[:firstIndex])
		releases = releases[firstIndex:]
	}
	return releases, nil
}

func GetSortedReleases(url string, pattern any, num int) ([]string, error) {
	return GetSortedReleasesFunc(url, pattern, num, strings.Compare)
}

func GetReverseReleases(url string, pattern any, num int) (iter.Seq[string], int, error) {
	page, err := CapturePage(url)
	if err != nil {
		return nil, 0, err
	}
	releaseRe, err := toRegexp(pattern)
	if err != nil {
		return nil, 0, err
	}
	matches := releaseRe.FindAllStringSubmatch(page, -1)
	if num >= 0 {
		numReturns := min(len(matches), num)
		firstIndex := len(matches) - numReturns
		// Allow GC to free unused matches
		clear(matches[:firstIndex])
		matches = matches[firstIndex:]
	}
	return func(yield func(string) bool) {
		for _, match := range slices.Backward(matches) {
			if !yield(match[1]) {
				return
			}
		}
	}, len(matches), nil
}

func GetBasicReleases(url string, pattern any, num int) (iter.Seq[string], int, error) {
	page, err := CapturePage(url)
	if err != nil {
		return nil, 0, err
	}
	releaseRe, err := toRegexp(pattern)
	if err != nil {
		return nil, 0, err
	}

	matches := releaseRe.FindAllStringSubmatch(page, num)
	return func(yield func(string) bool) {
		for _, match := range matches {
			if !yield(match[1]) {
				return
			}
		}
	}, len(matches), nil
}

func toRegexp(pattern any) (*regexp.Regexp, error) {
	switch p := pattern.(type) {
	case string:
		return regexp.Compile(p)
	case *regexp.Regexp:
		return p, nil
	default:
		return nil, fmt.Errorf("invalid pattern type %T", p)
	}
}

func IntegerCompare(a, b string) int {
	aInt, aErr := strconv.Atoi(a)
	bInt, bErr := strconv.Atoi(b)
	if aErr != nil && bErr != nil {
		return cmp.Compare(aInt, bInt)
	} else if aErr != nil {
		return -1
	} else if bErr != nil {
		return 1
	}
	return strings.Compare(a, b)
}

func SemverCompare(a, b string) int {
	aSemver, aErr := version.NewVersion(a)
	bSemver, bErr := version.NewVersion(b)
	if aErr != nil && bErr != nil {
		return aSemver.Compare(bSemver)
	} else if aErr != nil {
		return -1
	} else if bErr != nil {
		return 1
	}
	return IntegerCompare(a, b)
}
