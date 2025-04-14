package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const cachyOSMirror = "https://mirror.cachyos.org/ISO/"

var CachyOS = OS{
	Name:           "cachyos",
	PrettyName:     "CachyOS",
	Homepage:       "https://cachyos.org/",
	Description:    "Designed to deliver lightning-fast speeds and stability, ensuring a smooth and enjoyable computing experience every time you use it.",
	ConfigFunction: createCachyOSConfigs,
}

func createCachyOSConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	mirrors, err := getCachyOSEditionMirrors()
	if err != nil {
		return nil, err
	}
	ch, wg := getChannelsWith(len(mirrors))
	releaseRe := regexp.MustCompile(`href="([0-9]+)/"`)
	isoRe := regexp.MustCompile(`href="(cachyos-([^-]+)-linux-[0-9]+.iso)"`)

	for _, mirror := range mirrors {
		go func() {
			defer wg.Done()
			releases, numReleases, err := getBasicReleases(mirror, releaseRe, -1)
			if err != nil {
				errs <- Failure{Error: err}
			}
			wg.Add(numReleases)
			for release := range releases {
				mirror := mirror + release + "/"
				go func() {
					defer wg.Done()
					page, err := web.CapturePage(mirror)
					if err != nil {
						errs <- Failure{Release: release, Error: err}
						return
					}
					matches := isoRe.FindAllStringSubmatch(page, -1)
					wg.Add(len(matches))
					for _, match := range matches {
						url := mirror + match[1]
						edition := match[2]
						go func() {
							defer wg.Done()
							checksum, err := cs.SingleWhitespace(url + ".sha256")
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
		}()
	}
	return waitForConfigs(ch, wg), nil
}

func getCachyOSEditionMirrors() ([]string, error) {
	editionData, err := web.CapturePage(cachyOSMirror)
	if err != nil {
		return nil, err
	}
	editionRe := regexp.MustCompile(`href="(\w+)\/`)
	matches := editionRe.FindAllStringSubmatch(editionData, -1)

	mirrors := make([]string, len(matches))
	for i, match := range matches {
		mirrors[i] = cachyOSMirror + match[1] + "/"
	}
	return mirrors, nil
}
