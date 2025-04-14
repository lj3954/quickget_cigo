package os

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	endlessDlMirror   = "https://images-dl.endlessm.com/release/"
	endlessDataMirror = "https://mirror.leitecastro.com/endless/release/"
	endlessReleaseRe  = `href="(\d+(?:.\d+){2})\/"`
)

var EndlessOS = OS{
	Name:           "endless",
	PrettyName:     "Endless OS",
	Homepage:       "https://endlessos.org/",
	Description:    "Completely Free, User-Friendly Operating System Packed with Educational Tools, Games, and More.",
	ConfigFunction: createEndlessOSConfigs,
}

func createEndlessOSConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, numReleases, err := getBasicReleases(endlessDataMirror, endlessReleaseRe, -1)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannelsWith(numReleases)
	editionRe := regexp.MustCompile(`href="([^./]+)`)
	isoRe := regexp.MustCompile(`href="(eos-eos[\d.]+-amd64-amd64.[-\d]+.[^.]+.iso)"`)

	for release := range releases {
		mirror := endlessDataMirror + release + "/eos-amd64-amd64/"
		go func() {
			defer wg.Done()
			editions, err := getEndlessEditions(mirror, editionRe)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}

			wg.Add(len(editions))
			for _, edition := range editions {
				mirror := mirror + edition + "/"
				go func() {
					defer wg.Done()
					page, err := web.CapturePage(mirror)
					if err != nil {
						errs <- Failure{Release: release, Edition: edition, Error: err}
						return
					}
					isoMatch := isoRe.FindStringSubmatch(page)
					if isoMatch == nil {
						errs <- Failure{Release: release, Edition: edition, Error: errors.New("No ISO found")}
						return
					}
					iso := isoMatch[1]
					url := fmt.Sprintf("%s%s/eos-amd64-amd64/%s/%s", endlessDlMirror, release, edition, iso)

					checksumUrl := url + ".sha256"
					checksum, err := cs.SingleWhitespace(checksumUrl)
					if err != nil {
						csErrs <- Failure{Release: release, Edition: edition, Error: err}
					}
					ch <- Config{
						Release: release,
						Edition: edition,
						ISO: []Source{
							urlChecksumSource(url, checksum),
						},
					}
				}()
			}
		}()
	}

	return waitForConfigs(ch, wg), nil
}

func getEndlessEditions(url string, editionRe *regexp.Regexp) ([]string, error) {
	page, err := web.CapturePage(url)
	if err != nil {
		return nil, err
	}
	matches := editionRe.FindAllStringSubmatch(page, -1)

	releases := make([]string, len(matches))
	for i, match := range matches {
		releases[i] = match[1]
	}
	return releases, nil
}
