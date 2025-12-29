package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	mxlinuxMirror    = "https://sourceforge.net/projects/mx-linux/files/Final/"
	mxlinuxReleaseRe = `title="(\w+)" class="folder`
)

var MXLinux = OS{
	Name:           "mxlinux",
	PrettyName:     "MX Linux",
	Homepage:       "https://mxlinux.org/",
	Description:    "Designed to combine elegant and efficient desktops with high stability and solid performance.",
	ConfigFunction: createMXLinuxConfigs,
}

func createMXLinuxConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	editions, _, err := getBasicReleases(mxlinuxMirror, mxlinuxReleaseRe, -1)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannels()
	isoRe := regexp.MustCompile(`"name":"(MX-([\d\.]+)(_\w+)?_x64.iso)"`)

	for edition := range editions {
		mirror := mxlinuxMirror + edition + "/"
		wg.Go(func() {
			page, err := web.CapturePage(mirror)
			if err != nil {
				errs <- Failure{Release: edition, Error: err}
				return
			}
			matches := isoRe.FindAllStringSubmatch(page, -1)
			for _, match := range matches {
				iso, release, edition := match[1], match[2], match[3]
				if edition == "" {
					edition = "XFCE"
				} else {
					edition = edition[1:]
				}

				mirror := mirror + iso
				url := mirror + "/download"
				checksumUrl := mirror + ".sha256/download"

				wg.Go(func() {
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
				})
			}
		})
	}
	return waitForConfigs(ch, wg), nil
}
