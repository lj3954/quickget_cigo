package os

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/quickemu-project/quickget_configs/internal/cs"
)

const (
	antiXMirror    = "https://sourceforge.net/projects/antix-linux/files/Final/"
	antiXReleaseRe = `"name":"antiX-([0-9.]+)"`
)

type AntiX struct{}

func (AntiX) Data() OSData {
	return OSData{
		Name:        "antix",
		PrettyName:  "antiX",
		Homepage:    "https://antixlinux.com/",
		Description: "Fast, lightweight and easy to install systemd-free linux live CD distribution based on Debian Stable for Intel-AMD x86 compatible systems.",
	}
}

func (AntiX) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	releases, err := getBasicReleases(antiXMirror, antiXReleaseRe, 3)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannels()
	isoRe := regexp.MustCompile(`"name":"(antiX-[0-9.]+(?:-runit)?(?:-[^_]+)?_x64-([^.]+).iso)".*?"download_url":"(.*?)"`)

	for release := range releases {
		mirror := fmt.Sprintf("%santiX-%s/", antiXMirror, release)
		checksumUrl := mirror + "README.txt/download"
		createAntiXConfigs(ch, errs, csErrs, &wg, release, mirror, checksumUrl, isoRe, "-sysv")

		runitMirror := fmt.Sprintf("%srunit-antiX-%s/", mirror, release)
		runitChecksumUrl := runitMirror + "README2.txt/download"
		createAntiXConfigs(ch, errs, csErrs, &wg, release, runitMirror, runitChecksumUrl, isoRe, "-runit")
	}

	return waitForConfigs(ch, &wg), nil
}

func createAntiXChecksums(url string) (map[string]string, error) {
	page, err := capturePage(url)
	if err != nil {
		return nil, err
	}
	data := strings.SplitN(page, "sha256:", 2)
	if len(data) != 2 {
		return nil, errors.New("Could not find antiX 'sha256' separator")
	}
	return cs.Whitespace{}.BuildWithData(data[1]), nil
}

func createAntiXConfigs(ch chan Config, errs, csErrs chan Failure, wg *sync.WaitGroup, release, url, checksumUrl string, isoRe *regexp.Regexp, editionSuffix string) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		page, err := capturePage(url)
		if err != nil {
			errs <- Failure{Release: release, Error: err}
			return
		}
		checksums, err := createAntiXChecksums(checksumUrl)
		if err != nil {
			csErrs <- Failure{Release: release, Error: err}
		}
		for _, match := range isoRe.FindAllStringSubmatch(page, -1) {
			checksum, url := checksums[match[1]], match[3]
			ch <- Config{
				Release: release,
				Edition: match[2] + editionSuffix,
				ISO: []Source{
					urlChecksumSource(url, checksum),
				},
			}
		}
	}()
}
