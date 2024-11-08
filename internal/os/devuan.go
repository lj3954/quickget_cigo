package os

import "regexp"

const (
	devuanMirror    = "https://files.devuan.org/"
	devuanReleaseRe = `href="(devuan_[a-zA-Z]+/)"`
)

type Devuan struct{}

func (Devuan) Data() OSData {
	return OSData{
		Name:        "devuan",
		PrettyName:  "Devuan",
		Homepage:    "https://devuan.org/",
		Description: "Fork of Debian without systemd that allows users to reclaim control over their system by avoiding unnecessary entanglements and ensuring Init Freedom.",
	}
}

func (Devuan) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	releases, err := getBasicReleases(devuanMirror, devuanReleaseRe, -1)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannels()
	isoRe := regexp.MustCompile(`href="(devuan_[a-zA-Z]+_([0-9.]+)_amd64_desktop-live.iso)"`)
	csUrlRe := regexp.MustCompile(`href="(SHA[^.]+.txt)"`)

	wg.Add(len(releases))
	for _, urlSuffix := range releases {
		mirror := devuanMirror + urlSuffix + "desktop-live/"
		go func() {
			defer wg.Done()
			page, err := capturePage(mirror)
			if err != nil {
				errs <- Failure{Error: err}
				return
			}

			checksums := make(map[string]string)
			csUrlMatch := csUrlRe.FindStringSubmatch(page)
			if csUrlMatch != nil {
				checksumUrl := mirror + csUrlMatch[1]
				cs, err := buildChecksum(Whitespace{}, checksumUrl)
				if err != nil {
					csErrs <- Failure{Error: err}
				} else {
					checksums = cs
				}
			}

			for _, match := range isoRe.FindAllStringSubmatch(page, -1) {
				iso := match[1]
				url := mirror + iso
				checksum := checksums[iso]
				ch <- Config{
					Release: match[2],
					ISO: []Source{
						urlChecksumSource(url, checksum),
					},
				}
			}
		}()
	}

	return waitForConfigs(ch, &wg), nil
}
