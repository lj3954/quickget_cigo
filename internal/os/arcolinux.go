package os

import "regexp"

const ArcoLinuxMirror = "https://mirror.accum.se/mirror/arcolinux.info/iso/"

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
	releases, err := getArcoLinuxReleases()
	if err != nil {
		return nil, err
	}

	isoRe := regexp.MustCompile(`>(arco([^-]+)-[v0-9.]+-x86_64.iso)</a>`)
	csRegex := CustomRegex{
		Regex:      regexp.MustCompile(`>(arco([^-]+)-[v0-9.]+-x86_64.iso.sha256)</a>`),
		KeyIndex:   2,
		ValueIndex: 1,
	}
	ch, wg := getChannels()

	for _, release := range releases {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mirror := ArcoLinuxMirror + release + "/"
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
						cs, err := singleWhitespaceChecksum(mirror + checksumUrlExt)
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

func getArcoLinuxReleases() ([]string, error) {
	page, err := capturePage(ArcoLinuxMirror)
	if err != nil {
		return nil, err
	}
	releaseRe := regexp.MustCompile(`>(v[0-9.]+)/</a`)
	matches := releaseRe.FindAllStringSubmatch(page, -1)

	releases := make([]string, len(matches))
	for i := len(matches) - 1; i >= len(matches)-3 && i >= 0; i-- {
		releases[i] = matches[i][1]
	}
	return releases, nil
}
