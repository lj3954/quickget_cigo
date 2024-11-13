package os

import (
	"errors"
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
)

const (
	garudaMirror    = "https://iso.builds.garudalinux.org/iso/latest/garuda/"
	garudaEditionRe = `href="([^.]+)\/"`
)

type Garuda struct{}

func (Garuda) Data() OSData {
	return OSData{
		Name:        "garuda",
		PrettyName:  "Garuda Linux",
		Homepage:    "https://garudalinux.org/",
		Description: "Feature rich and easy to use Linux distribution.",
	}
}

func (Garuda) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	editions, err := getBasicReleases(garudaMirror, garudaEditionRe, -1)
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(`href="([^"]+.iso)"`)
	ch, wg := getChannels()

	release := "latest"
	for edition := range editions {
		mirror := garudaMirror + edition + "/"
		wg.Add(1)
		go func() {
			defer wg.Done()
			page, err := capturePage(mirror)
			if err != nil {
				errs <- Failure{Release: release, Edition: edition, Error: err}
				return
			}
			isoMatch := isoRe.FindStringSubmatch(page)
			if isoMatch == nil {
				errs <- Failure{Release: release, Edition: edition, Error: errors.New("No ISO found")}
				return
			}
			url := mirror + isoMatch[1]

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

	return waitForConfigs(ch, &wg), nil
}
