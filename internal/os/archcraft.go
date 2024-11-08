package os

import (
	"fmt"
	"regexp"
)

const (
	archcraftMirror    = "https://sourceforge.net/projects/archcraft/files/"
	archcraftReleaseRe = `"name":"v([^"]+)"`
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

func (Archcraft) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	releases, err := getBasicReleases(archcraftMirror, archcraftReleaseRe, 3)
	if err != nil {
		return nil, err
	}
	urlRe := regexp.MustCompile(`"name":"archcraft-.*?-x86_64.iso".*?"download_url":"([^"]+)".*?"name":"archcraft-.*?-x86_64.iso.sha256sum".*?"download_url":"([^"]+)"`)
	ch, wg := getChannels()
	for _, release := range releases {
		mirror := fmt.Sprintf("%sv%s/", archcraftMirror, release)
		wg.Add(1)
		go func() {
			defer wg.Done()
			page, err := capturePage(mirror)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			urls := urlRe.FindStringSubmatch(page)
			if len(urls) == 3 {
				checksum, err := singleWhitespaceChecksum(urls[2])
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

	return waitForConfigs(ch, &wg), nil
}
