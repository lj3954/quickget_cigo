package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	bodhiMirror    = "https://sourceforge.net/projects/bodhilinux/files/"
	bodhiReleaseRe = `"name":"([0-9]+.[0-9]+.[0-9]+)"`
)

type Bodhi struct{}

func (Bodhi) Data() OSData {
	return OSData{
		Name:        "bodhi",
		PrettyName:  "Bodhi",
		Homepage:    "https://www.bodhilinux.com/",
		Description: "Lightweight distribution featuring the fast & fully customizable Moksha Desktop.",
	}
}

func (Bodhi) CreateConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, numReleases, err := getBasicReleases(bodhiMirror, bodhiReleaseRe, 3)
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(`"name":"(bodhi-[0-9]+.[0-9]+.[0-9]+-64(-[^-.]+)?.iso)"`)
	ch, wg := getChannelsWith(numReleases)

	for release := range releases {
		mirror := bodhiMirror + release + "/"
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
				edition := "standard"
				if match[2] != "" {
					edition = match[2][1:]
				}
				url := mirror + match[1] + "/download"
				checksumUrl := mirror + match[1] + ".sha256/download"

				go func() {
					defer wg.Done()
					checksum, err := cs.SingleWhitespace(checksumUrl)
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

	return waitForConfigs(ch, wg), nil
}
