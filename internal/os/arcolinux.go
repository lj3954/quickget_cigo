package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
)

const (
	arcolinuxMirror    = "https://mirror.accum.se/mirror/arcolinux.info/iso/"
	arcolinuxReleaseRe = `>(v[0-9.]+)/</a`
)

type ArcoLinux struct{}

func (ArcoLinux) Data() OSData {
	return OSData{
		Name:        "arcolinux",
		PrettyName:  "ArcoLinux",
		Homepage:    "https://arcolinux.com/",
		Description: "It's all about becoming an expert in Linux.",
	}
}

func (ArcoLinux) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	releases, err := getBasicReleases(arcolinuxMirror, arcolinuxReleaseRe, -1)
	if err != nil {
		return nil, err
	}

	isoRe := regexp.MustCompile(`>(arco([^-]+)-[v0-9.]+-x86_64.iso)</a>`)
	csRegex := cs.CustomRegex{
		Regex:      regexp.MustCompile(`>(arco([^-]+)-[v0-9.]+-x86_64.iso.sha256)</a>`),
		KeyIndex:   2,
		ValueIndex: 1,
	}
	ch, wg := getChannels()

	for release := range releases {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mirror := arcolinuxMirror + release + "/"
			page, err := capturePage(mirror)
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

	return waitForConfigs(ch, &wg), nil
}
