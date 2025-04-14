package os

import (
	"regexp"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	porteusMirror    = "https://mirrors.dotsrc.org/porteus/x86_64/"
	porteusReleaseRe = `href="(Porteus-v[\d\.]+/)"`
	porteusIsoRe     = `href="(Porteus-([^-]+)-(.*?)-x86_64.iso)"`
)

var Porteus = OS{
	Name:           "porteus",
	PrettyName:     "Porteus",
	Homepage:       "http://www.porteus.org/",
	Description:    "Complete linux operating system that is optimized to run from CD, USB flash drive, hard drive, or other bootable storage media.",
	ConfigFunction: createPorteusConfigs,
}

func createPorteusConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, numReleases, err := getReverseReleases(porteusMirror, porteusReleaseRe, 3)
	if err != nil {
		return nil, err
	}

	isoRe := regexp.MustCompile(porteusIsoRe)
	ch, wg := getChannelsWith(numReleases)
	for release := range releases {
		go func() {
			defer wg.Done()
			url := porteusMirror + release
			page, err := web.CapturePage(url)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}

			matches := isoRe.FindAllStringSubmatch(page, -1)
			if len(matches) == 0 {
				return
			}
			checksums, err := cs.Build(cs.Whitespace{}, url+"sha256sums.txt")
			if err != nil {
				csErrs <- Failure{Release: release, Error: err}
			}

			for _, match := range matches {
				iso := match[1]
				checksum := checksums[iso]
				ch <- Config{
					Release: match[3],
					Edition: strings.ToLower(match[2]),
					ISO: []Source{
						urlChecksumSource(url+iso, checksum),
					},
				}
			}
		}()
	}

	return waitForConfigs(ch, wg), nil
}
