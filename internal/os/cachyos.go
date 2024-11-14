package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
)

const cachyOSMirror = "https://mirror.cachyos.org/ISO/"

type CachyOS struct{}

func (CachyOS) Data() OSData {
	return OSData{
		Name:        "cachyos",
		PrettyName:  "CachyOS",
		Homepage:    "https://cachyos.org/",
		Description: "Designed to deliver lightning-fast speeds and stability, ensuring a smooth and enjoyable computing experience every time you use it.",
	}
}

func (CachyOS) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	mirrors, err := getCachyOSEditionMirrors()
	if err != nil {
		return nil, err
	}
	ch, wg := getChannels()
	releaseRe := regexp.MustCompile(`href="([0-9]+)/"`)
	isoRe := regexp.MustCompile(`href="(cachyos-([^-]+)-linux-[0-9]+.iso)"`)

	for _, mirror := range mirrors {
		wg.Add(1)
		go func() {
			defer wg.Done()
			releases, err := getBasicReleases(mirror, releaseRe, -1)
			if err != nil {
				errs <- Failure{Error: err}
			}
			for release := range releases {
				mirror := mirror + release + "/"
				wg.Add(1)
				go func() {
					defer wg.Done()
					page, err := capturePage(mirror)
					if err != nil {
						errs <- Failure{Release: release, Error: err}
						return
					}
					for _, match := range isoRe.FindAllStringSubmatch(page, -1) {
						url := mirror + match[1]
						edition := match[2]
						wg.Add(1)
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
	return waitForConfigs(ch, &wg), nil
}

func getCachyOSEditionMirrors() ([]string, error) {
	editionData, err := capturePage(cachyOSMirror)
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
