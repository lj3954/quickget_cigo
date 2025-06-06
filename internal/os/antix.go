package os

import (
	"errors"
	"fmt"
	"iter"
	"regexp"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	antiXMirror    = "https://sourceforge.net/projects/antix-linux/files/Final/"
	antiXReleaseRe = `"name":"antiX-([0-9.]+)"`
)

var AntiX = OS{
	Name:           "antix",
	PrettyName:     "antiX",
	Homepage:       "https://antixlinux.com/",
	Description:    "Fast, lightweight and easy to install systemd-free linux live CD distribution based on Debian Stable for Intel-AMD x86 compatible systems.",
	ConfigFunction: createAntiXConfigs,
}

func createAntiXConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, numReleases, err := getBasicReleases(antiXMirror, antiXReleaseRe, 3)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannelsWith(numReleases * 2)
	isoRe := regexp.MustCompile(`"name":"(antiX-[0-9.]+(?:-runit)?(?:-[^_]+)?_x64-([^.]+).iso)".*?"download_url":"(.*?)"`)

	var addConfigs = func(release string, mirror string, checksumUrl string, editionSuffix string) {
		defer wg.Done()
		configs, csErr, err := createFinalAntiXConfigs(release, mirror, checksumUrl, isoRe, editionSuffix)
		if err != nil {
			errs <- Failure{Release: release, Error: err}
			return
		}
		if csErr != nil {
			csErrs <- Failure{Release: release, Error: csErr}
		}
		for config := range configs {
			ch <- config
		}
	}

	for release := range releases {
		mirror := fmt.Sprintf("%santiX-%s/", antiXMirror, release)
		checksumUrl := mirror + "README.txt/download"
		go addConfigs(release, mirror, checksumUrl, "-sysv")

		runitMirror := fmt.Sprintf("%srunit-antiX-%s/", mirror, release)
		runitChecksumUrl := runitMirror + "README2.txt/download"
		go addConfigs(release, runitMirror, runitChecksumUrl, "-runit")
	}

	return waitForConfigs(ch, wg), nil
}

func createAntiXChecksums(url string) (map[string]string, error) {
	page, err := web.CapturePage(url)
	if err != nil {
		return nil, err
	}
	data := strings.SplitN(page, "sha256:", 2)
	if len(data) != 2 {
		return nil, errors.New("Could not find antiX 'sha256' separator")
	}
	return cs.Whitespace.BuildWithData(data[1]), nil
}

func createFinalAntiXConfigs(release, url, checksumUrl string, isoRe *regexp.Regexp, editionSuffix string) (configs iter.Seq[Config], csErr error, err error) {
	page, err := web.CapturePage(url)
	if err != nil {
		return
	}
	checksums, err := createAntiXChecksums(checksumUrl)
	if err != nil {
		csErr = err
	}

	return func(yield func(Config) bool) {
		for _, match := range isoRe.FindAllStringSubmatch(page, -1) {
			checksum, url := checksums[match[1]], match[3]
			config := Config{
				Release: release,
				Edition: match[2] + editionSuffix,
				ISO: []Source{
					urlChecksumSource(url, checksum),
				},
			}

			if !yield(config) {
				return
			}
		}
	}, csErr, nil
}
