package os

import (
	"errors"
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
	"github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

const (
	freedosMirror    = "https://www.ibiblio.org/pub/micro/pc-stuff/freedos/files/distributions/"
	freedosReleaseRe = `href="(\d+\.\d+)/"`
)

var FreeDOS = OS{
	Name:           "freedos",
	PrettyName:     "FreeDOS",
	Homepage:       "https://www.freedos.org/",
	Description:    "DOS-compatible operating system that you can use to play classic DOS games, run legacy business software, or develop embedded systems.",
	ConfigFunction: createFreeDOSConfigs,
}

func createFreeDOSConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, numReleases, err := getBasicReleases(freedosMirror, freedosReleaseRe, -1)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannelsWith(numReleases)
	isoRe := regexp.MustCompile(`href="(FD\d+-?(.*?CD)\.(iso|zip))"`)
	checksumRe := regexp.MustCompile(`FD\d+.sha|verify.txt`)

	for release := range releases {
		mirror := freedosMirror + release + "/official/"
		go func() {
			defer wg.Done()
			page, err := web.CapturePage(mirror)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}

			checksums, err := getFreeDOSChecksums(mirror, page, checksumRe)
			if err != nil {
				csErrs <- Failure{Release: release, Error: err}
			}

			for _, match := range isoRe.FindAllStringSubmatch(page, -1) {
				iso, edition, filetype := match[1], match[2], match[3]
				url := mirror + iso
				checksum := checksums[iso]

				var archiveFormat ArchiveFormat
				if filetype == "zip" {
					archiveFormat = quickgetdata.Zip
				}
				ch <- Config{
					GuestOS: quickgetdata.FreeDOS,
					Release: release,
					Edition: edition,
					ISO: []Source{
						webSource(url, checksum, archiveFormat, ""),
					},
				}
			}
		}()
	}
	return waitForConfigs(ch, wg), nil
}

func getFreeDOSChecksums(url, page string, checksumRe *regexp.Regexp) (map[string]string, error) {
	csUrlMatch := checksumRe.FindString(page)
	if csUrlMatch == "" {
		return nil, errors.New("Could not find Checksum URL")
	}
	return cs.Build(cs.Whitespace{}, url+csUrlMatch)
}
