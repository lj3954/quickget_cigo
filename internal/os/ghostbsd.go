package os

import (
	"regexp"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/web"
	"github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

const (
	ghostbsdMirror    = "https://download.ghostbsd.org/releases/amd64/"
	ghostbsdReleaseRe = `href="(latest|[\d\.]+)\/"`
)

var GhostBSD = OS{
	Name:           "ghostbsd",
	PrettyName:     "GhostBSD",
	Homepage:       "https://www.ghostbsd.org/",
	Description:    "Simple, elegant desktop BSD Operating System.",
	ConfigFunction: createGhostBSDConfigs,
}

func createGhostBSDConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, numReleases, err := getReverseReleases(ghostbsdMirror, ghostbsdReleaseRe, 4)
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(`href="(GhostBSD-[\d\.]+(-[\w]+)?.iso)"`)
	ch, wg := getChannelsWith(numReleases)

	for release := range releases {
		mirror := ghostbsdMirror + release + "/"
		go func() {
			defer wg.Done()
			page, err := web.CapturePage(mirror)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			for _, match := range isoRe.FindAllStringSubmatch(page, -1) {
				edition := match[2]
				if edition == "" {
					edition = "MATE"
				} else {
					edition = edition[1:]
				}

				iso := match[1]
				url := mirror + iso
				checksumUrl := url + ".sha256"

				wg.Add(1)
				go func() {
					defer wg.Done()
					checksum, err := web.CapturePage(checksumUrl)
					if err != nil {
						csErrs <- Failure{Release: release, Edition: edition, Error: err}
					}
					checksum = checksum[strings.Index(checksum, "=")+1:]

					ch <- Config{
						Release: release,
						Edition: edition,
						GuestOS: quickgetdata.GhostBSD,
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
