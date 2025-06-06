package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	voidMirror    = "https://repo-default.voidlinux.org/live/"
	voidReleaseRe = `href="(\d{8})\/"`
	voidIsoRe     = `href="(void-live-(aarch64|x86_64)(-musl)?-\d{8}-(.*?)\.iso)"`
)

var Void = OS{
	Name:           "void",
	PrettyName:     "Void Linux",
	Homepage:       "https://voidlinux.org/",
	Description:    "General purpose operating system. Its package system allows you to quickly install, update and remove software; software is provided in binary packages or can be built directly from sources.",
	ConfigFunction: createVoidConfigs,
}

func createVoidConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, numReleases, err := getBasicReleases(voidMirror, voidReleaseRe, -1)
	if err != nil {
		return nil, err
	}

	isoRe := regexp.MustCompile(voidIsoRe)
	ch, wg := getChannelsWith(numReleases)
	for release := range releases {
		go func() {
			defer wg.Done()
			url := voidMirror + release + "/"
			checksums, err := cs.Build(cs.Sha256Regex, url+"sha256sum.txt")
			if err != nil {
				csErrs <- Failure{Release: release, Error: err}
			}

			page, err := web.CapturePage(url)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			for _, match := range isoRe.FindAllStringSubmatch(page, -1) {
				iso := match[1]
				url := url + iso
				checksum := checksums[iso]
				arch := Arch(match[2])
				edition := match[4] + match[3]

				ch <- Config{
					Release: release,
					Edition: edition,
					Arch:    arch,
					ISO: []Source{
						urlChecksumSource(url, checksum),
					},
				}
			}
		}()
	}

	return waitForConfigs(ch, wg), nil
}
