package os

import "regexp"

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

func (Bodhi) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	releases, err := getBasicReleases(bodhiMirror, bodhiReleaseRe, 3)
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(`"name":"(bodhi-[0-9]+.[0-9]+.[0-9]+-64(-[^-.]+)?.iso)"`)
	ch, wg := getChannels()

	for _, release := range releases {
		mirror := bodhiMirror + release + "/"
		wg.Add(1)
		go func() {
			defer wg.Done()
			page, err := capturePage(mirror)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			for _, match := range isoRe.FindAllStringSubmatch(page, -1) {
				edition := "standard"
				if match[2] != "" {
					edition = match[2][1:]
				}
				url := mirror + match[1] + "/download"
				checksumUrl := mirror + match[1] + ".sha256/download"

				wg.Add(1)
				go func() {
					defer wg.Done()
					checksum, err := singleWhitespaceChecksum(checksumUrl)
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

	return waitForConfigs(ch, &wg), nil
}
