package os

import (
	"fmt"
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
)

const (
	mageiaMirror    = "https://mirror.math.princeton.edu/pub/mageia/iso/"
	mageiaReleaseRe = `href="(\d+)\/"`
)

type Mageia struct{}

func (Mageia) Data() OSData {
	return OSData{
		Name:        "mageia",
		PrettyName:  "Mageia",
		Homepage:    "https://www.mageia.org/",
		Description: "Stable, secure operating system for desktop & server.",
	}
}

func (Mageia) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	releases, numReleases, err := getBasicReleases(mageiaMirror, mageiaReleaseRe, -1)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannelsWith(numReleases)
	editionRe := regexp.MustCompile(`href="Mageia-\d+-Live-([^-]+)-x86_64`)

	for release := range releases {
		mirror := mageiaMirror + release + "/"
		go func() {
			defer wg.Done()
			editions, numEditions, err := getBasicReleases(mirror, editionRe, -1)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			wg.Add(numEditions)
			for edition := range editions {
				isoText := fmt.Sprintf("Mageia-%s-Live-%s-x86_64", release, edition)
				go func() {
					defer wg.Done()
					url := mirror + isoText + "/" + isoText + ".iso"
					checksum, err := cs.SingleWhitespace(url + ".sha512")
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
