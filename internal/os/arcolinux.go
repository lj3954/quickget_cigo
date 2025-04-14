package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	arcoLinuxMirror    = "https://ant.seedhost.eu/arcolinux/iso/"
	arcoLinuxReleaseRe = `href='./(v[0-9\.]+)'`
	arcoLinuxIsoRe     = `>(arco([^-]+)-[v0-9.]+-x86_64.iso)</a>`
	arcoLinuxCsRe      = `>(arco([^-]+)-[v0-9.]+-x86_64.iso.sha256)</a>`
)

var ArcoLinux = OS{
	Name:           "arcolinux",
	PrettyName:     "ArcoLinux",
	Homepage:       "https://arcolinux.com/",
	Description:    "It's all about becoming an expert in Linux.",
	ConfigFunction: createArcoLinuxConfigs,
}

func createArcoLinuxConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, numReleases, err := getBasicReleases(arcoLinuxMirror, arcoLinuxReleaseRe, 3)
	if err != nil {
		return nil, err
	}

	isoRe := regexp.MustCompile(arcoLinuxIsoRe)
	csRegex := cs.CustomRegex{
		Regex:      regexp.MustCompile(arcoLinuxCsRe),
		KeyIndex:   2,
		ValueIndex: 1,
	}
	ch, wg := getChannelsWith(numReleases)

	for release := range releases {
		go func() {
			defer wg.Done()
			mirror := arcoLinuxMirror + release + "/"
			page, err := web.CapturePage(mirror)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			checksums := csRegex.BuildWithData(page)

			matches := isoRe.FindAllStringSubmatch(page, -1)
			for _, match := range matches {
				url := mirror + match[1]
				edition := match[2]
				checksumUrlExt, checksumUrlExists := checksums[match[2]]
				wg.Add(1)
				go func() {
					defer wg.Done()
					var checksum string
					if checksumUrlExists {
						cs, err := cs.SingleWhitespace(mirror + checksumUrlExt)
						if err != nil {
							csErrs <- Failure{Release: release, Edition: edition, Error: err}
						} else {
							checksum = cs
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
