package os

import (
	"regexp"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const gentooMirror = "https://distfiles.gentoo.org/releases/"

var Gentoo = OS{
	Name:           "gentoo",
	PrettyName:     "Gentoo",
	Homepage:       "https://www.gentoo.org/",
	Description:    "Highly flexible, source-based Linux distribution.",
	ConfigFunction: createGentooConfigs,
}

func createGentooConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	architectures := [...]string{"amd64", "arm64"}
	isoRe := regexp.MustCompile(`\d{8}T\d{6}Z\/(admincd|install|livegui).*?.iso`)
	ch, wg := getChannelsWith(len(architectures))

	release := "latest"
	for _, arch := range architectures {
		mirror := gentooMirror + arch + "/autobuilds/"
		go func() {
			defer wg.Done()
			page, err := web.CapturePage(mirror + "latest-iso.txt")
			if err != nil {
				errs <- Failure{Release: release, Arch: Arch(arch), Error: err}
				return
			}
			matches := isoRe.FindAllStringSubmatch(page, -1)
			wg.Add(len(matches))
			for _, match := range matches {
				edition := match[1]
				if edition == "install" {
					edition = "minimal"
				}
				url := mirror + match[0]
				checksumUrl := url + ".sha256"

				go func() {
					defer wg.Done()
					checksumPage, err := web.CapturePage(checksumUrl)
					if err != nil {
						csErrs <- Failure{Release: release, Edition: edition, Arch: Arch(arch), Error: err}
					}
					var checksum string
					for _, line := range strings.Split(checksumPage, "\n") {
						if strings.Contains(line, "iso") {
							cs, err := cs.BuildSingleWhitespace(line)
							if err != nil {
								csErrs <- Failure{Release: release, Edition: edition, Arch: Arch(arch), Error: err}
							}
							checksum = cs
							break
						}
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
