package utils

import (
	"cmp"
	"fmt"
	"iter"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/hashicorp/go-version"
	"github.com/quickemu-project/quickget_configs/internal/data"
	"github.com/quickemu-project/quickget_configs/internal/web"
	"github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

type Config = quickgetdata.Config
type Failure = data.Failure
type OSData = quickgetdata.OSData

type Distro interface {
	Data() OSData
	CreateConfigs(chan<- data.Failure, chan<- data.Failure) ([]Config, error)
}

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

func GetSortedReleasesFunc(url string, pattern any, num int, cmp func(a, b string) int) ([]string, error) {
	page, err := web.CapturePage(url)
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
	page, err := web.CapturePage(url)
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
	page, err := web.CapturePage(url)
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
	if aErr == nil && bErr == nil {
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
	if aErr == nil && bErr == nil {
		return aSemver.Compare(bSemver)
	} else if aErr != nil {
		return -1
	} else if bErr != nil {
		return 1
	}
	return IntegerCompare(a, b)
}
