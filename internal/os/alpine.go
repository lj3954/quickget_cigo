package os

import (
	"fmt"
	"regexp"
)

const AlpineMirror = "https://dl-cdn.alpinelinux.org/alpine/"

type Alpine struct{}

func (Alpine) Data() OSData {
	return OSData{
		Name:        "alpine",
		PrettyName:  "Alpine Linux",
		Homepage:    "https://alpinelinux.org/",
		Description: "Security-oriented, lightweight Linux distribution based on musl libc and busybox.",
	}
}

func (Alpine) CreateConfigs() ([]Config, error) {
	releases, err := getAlpineReleases()
	if err != nil {
		return nil, err
	}
	ch, errs, wg := getChannels()
	isoRe := regexp.MustCompile(`(?s)iso: (alpine-virt-[0-9]+\.[0-9]+.*?.iso).*? sha256: ([0-9a-f]+)`)

	architectures := [2]Arch{x86_64, aarch64}
	for _, release := range releases {
		for _, arch := range architectures {
			mirror := fmt.Sprintf("%s%s/releases/%s/", AlpineMirror, release, arch)
			releaseUrl := mirror + "latest-releases.yaml"
			wg.Add(1)
			go func() {
				defer wg.Done()
				page, err := capturePage(releaseUrl)
				if err != nil {
					errs <- err
					return
				}

				if slice := isoRe.FindStringSubmatch(page); len(slice) > 0 {
					iso, checksum := slice[1], slice[2]
					url := mirror + iso
					ch <- Config{
						Release: release,
						Arch:    arch,
						ISO: []Source{
							urlChecksumSource(url, checksum),
						},
					}
				}
			}()
		}
	}

	return waitForConfigs(ch, errs, &wg), nil
}

func getAlpineReleases() ([]string, error) {
	page, err := capturePage(AlpineMirror)
	if err != nil {
		return nil, err
	}
	releaseRe := regexp.MustCompile(`<a href="(v[0-9]+\.[0-9]+)/"`)
	matches := releaseRe.FindAllStringSubmatch(page, -1)

	releases := make([]string, len(matches))
	for i, match := range matches {
		releases[i] = match[1]
	}

	return releases, nil
}