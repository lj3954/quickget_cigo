package os

import (
	"fmt"
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	archcraftMirror    = "https://sourceforge.net/projects/archcraft/files/"
	archcraftReleaseRe = `"name":"v([^"]+)"`
)

var archCraft = OS{
	Name:           "archcraft",
	PrettyName:     "Archcraft",
	Homepage:       "https://archcraft.io/",
	Description:    "Yet another minimal Linux distribution, based on Arch Linux.",
	ConfigFunction: createArchcraftConfigs,
}

func createArchcraftConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, numReleases, err := getBasicReleases(archcraftMirror, archcraftReleaseRe, 3)
	if err != nil {
		return nil, err
	}
	urlRe := regexp.MustCompile(`"name":"archcraft-.*?-x86_64.iso".*?"download_url":"([^"]+)".*?"name":"archcraft-.*?-x86_64.iso.sha256sum".*?"download_url":"([^"]+)"`)
	ch, wg := getChannelsWith(numReleases)

	for release := range releases {
		mirror := fmt.Sprintf("%sv%s/", archcraftMirror, release)
		go func() {
			defer wg.Done()
			page, err := web.CapturePage(mirror)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			urls := urlRe.FindStringSubmatch(page)
			if len(urls) == 3 {
				checksum, err := cs.SingleWhitespace(urls[2])
				if err != nil {
					csErrs <- Failure{Release: release, Error: err}
				}
				ch <- Config{
					Release: release,
					ISO: []Source{
						urlChecksumSource(urls[1], checksum),
					},
				}
			}
		}()
	}

	return waitForConfigs(ch, wg), nil
}
