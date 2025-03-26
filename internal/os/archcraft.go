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
	archcraftUrlRe     = `"name":"archcraft-.*?-x86_64.iso".*?"download_url":"([^"]+)".*?"name":"archcraft-.*?-x86_64.iso.sha256sum".*?"download_url":"([^"]+)"`
)

type Archcraft struct{}

func (Archcraft) Data() OSData {
	return OSData{
		Name:        "archcraft",
		PrettyName:  "Archcraft",
		Homepage:    "https://archcraft.io/",
		Description: "Yet another minimal Linux distribution, based on Arch Linux.",
	}
}

func (Archcraft) CreateConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, numReleases, err := getBasicReleases(archcraftMirror, archcraftReleaseRe, 3)
	if err != nil {
		return nil, err
	}
	urlRe := regexp.MustCompile(archcraftUrlRe)
	ch, wg := getChannelsWith(numReleases)

	for release := range releases {
		mirror := fmt.Sprintf("%sv%s/", archcraftMirror, release)
		go func() {
			defer wg.Done()
			config, csErr, err := getArchcraftConfig(release, mirror, urlRe)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			if csErr != nil {
				csErrs <- Failure{Release: release, Error: csErr}
			}
			ch <- *config
		}()
	}

	return waitForConfigs(ch, wg), nil
}

func getArchcraftConfig(release string, mirror string, urlRe *regexp.Regexp) (config *Config, csErr error, e error) {
	page, err := web.CapturePage(mirror)
	if err != nil {
		return nil, nil, err
	}
	urls := urlRe.FindStringSubmatch(page)
	if len(urls) == 3 {
		checksum, err := cs.SingleWhitespace(urls[2])
		if err != nil {
			csErr = err
		}
		config = &Config{
			Release: release,
			ISO: []Source{
				urlChecksumSource(urls[1], checksum),
			},
		}
	}

	return
}
