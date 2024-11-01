package os

import "regexp"

const BodhiMirror = "https://sourceforge.net/projects/bodhilinux/files/"

type Bodhi struct{}

func (Bodhi) Data() OSData {
	return OSData{
		Name:        "bodhi",
		PrettyName:  "Bodhi",
		Homepage:    "https://www.bodhilinux.com/",
		Description: "Lightweight distribution featuring the fast & fully customizable Moksha Desktop.",
	}
}

func (Bodhi) CreateConfigs() ([]Config, error) {
	releases, err := getBodhiReleases()
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(`"name":"(bodhi-[0-9]+.[0-9]+.[0-9]+-64(-[^-.]+)?.iso)"`)
	ch, errs, wg := getChannels()

	for _, release := range releases {
		mirror := BodhiMirror + release + "/"
		wg.Add(1)
		go func() {
			defer wg.Done()
			page, err := capturePage(mirror)
			if err != nil {
				errs <- err
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
						errs <- err
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

	return waitForConfigs(ch, errs, &wg), nil
}

func getBodhiReleases() ([]string, error) {
	page, err := capturePage(BodhiMirror)
	if err != nil {
		return nil, err
	}
	releaseRe := regexp.MustCompile(`"name":"([0-9]+.[0-9]+.[0-9]+)"`)
	matches := releaseRe.FindAllStringSubmatch(page, 3)

	releases := make([]string, len(matches))
	for i, match := range matches {
		releases[i] = match[1]
	}
	return releases, nil
}
