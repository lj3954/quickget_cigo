package os

import (
	"fmt"
	"regexp"
)

const ArchcraftMirror = "https://sourceforge.net/projects/archcraft/files/"

type Archcraft struct{}

func (Archcraft) Data() OSData {
	return OSData{
		Name:        "archcraft",
		PrettyName:  "Archcraft",
		Homepage:    "https://archcraft.io/",
		Description: "Yet another minimal Linux distribution, based on Arch Linux.",
	}
}

func (Archcraft) CreateConfigs(errs chan Failure) ([]Config, error) {
	releases, err := getArchcraftReleases()
	if err != nil {
		return nil, err
	}
	urlRe := regexp.MustCompile(`"name":"archcraft-.*?-x86_64.iso".*?"download_url":"([^"]+)".*?"name":"archcraft-.*?-x86_64.iso.sha256sum".*?"download_url":"([^"]+)"`)
	ch, wg := getChannels()
	for _, release := range releases {
		mirror := fmt.Sprintf("%sv%s/", ArchcraftMirror, release)
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
					errs <- Failure{Release: release, Error: err, Checksum: true}
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

func getArchcraftReleases() ([]string, error) {
	page, err := capturePage(ArchcraftMirror)
	if err != nil {
		return nil, err
	}
	releaseRe := regexp.MustCompile(`"name":"v([^"]+)"`)
	matches := releaseRe.FindAllStringSubmatch(page, 3)

	releases := make([]string, len(matches))
	for i, match := range matches {
		releases[i] = match[1]
	}
	return releases, nil
}
