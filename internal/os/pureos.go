package os

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	pureOsMirror    = "https://downloads.puri.sm/"
	pureOsReleaseRe = `href="(\d+)/"`
	pureOsEditionRe = `href="(\w.*?)/"`
	pureOsDateRe    = `href="(\d{4}-\d{2}-\d{2})/"`
	pureOsIsoRe     = `href="(pureos-([\d\.]+)-.*?-\d{8}_amd64.iso)"`
)

var PureOS = OS{
	Name:           "pureos",
	PrettyName:     "PureOS",
	Homepage:       "https://www.pureos.net/",
	Description:    "PureOS is a fully free/libre and open source GNU/Linux operating system, endorsed by the Free Software Foundation.",
	ConfigFunction: createPureOSConfigs,
}

func createPureOSConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, _, err := getBasicReleases(pureOsMirror, pureOsReleaseRe, -1)
	if err != nil {
		return nil, err
	}

	ch, wg := getChannels()
	editionRe := regexp.MustCompile(pureOsEditionRe)
	dateRe := regexp.MustCompile(pureOsDateRe)
	isoRe := regexp.MustCompile(pureOsIsoRe)

	for release := range releases {
		wg.Go(func() {
			url := pureOsMirror + release
			editions, _, err := getBasicReleases(url, editionRe, -1)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}

			for edition := range editions {
				wg.Go(func() {
					url := url + "/" + edition + "/"
					dates, _, err := getBasicReleases(url, dateRe, -1)
					if err != nil {
						errs <- Failure{Release: release, Error: err}
						return
					}
					dateSlice := slices.Collect(dates)
					if len(dateSlice) > 0 {
						date := slices.Max(slices.Collect(dates))
						url += date + "/"
					}

					page, err := web.CapturePage(url)
					if err != nil {
						errs <- Failure{Release: release, Error: err}
						return
					}

					isoMatch := isoRe.FindStringSubmatch(page)
					if len(isoMatch) == 0 {
						errs <- Failure{Release: release, Error: fmt.Errorf("No ISO found for %s", release)}
						return
					}
					iso := isoMatch[1]
					release := isoMatch[2]
					url += iso

					checksumUrl := strings.Replace(url, "iso", "checksums_sha256.txt", 1)
					checksums, err := cs.Build(cs.Whitespace, checksumUrl)
					if err != nil {
						csErrs <- Failure{Release: release, Error: err}
					}
					checksum := checksums["./"+iso]

					ch <- Config{
						Release: release,
						Edition: edition,
						ISO: []Source{
							urlChecksumSource(url, checksum),
						},
					}
				})
			}
		})
	}

	return waitForConfigs(ch, wg), nil
}
