package os

import (
	"fmt"
	"regexp"
)

const (
	alpineMirror    = "https://dl-cdn.alpinelinux.org/alpine/"
	alpineReleaseRe = `<a href="(v[0-9]+\.[0-9]+)/"`
)

type Alpine struct{}

func (Alpine) Data() OSData {
	return OSData{
		Name:        "alpine",
		PrettyName:  "Alpine Linux",
		Homepage:    "https://alpinelinux.org/",
		Description: "Security-oriented, lightweight Linux distribution based on musl libc and busybox.",
	}
}

func (Alpine) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	releases, err := getBasicReleases(alpineMirror, alpineReleaseRe, -1)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannels()
	isoRe := regexp.MustCompile(`(?s)iso: (alpine-virt-[0-9]+\.[0-9]+.*?.iso).*? sha256: ([0-9a-f]+)`)

	architectures := [2]Arch{x86_64, aarch64}
	for _, release := range releases {
		for _, arch := range architectures {
			mirror := fmt.Sprintf("%s%s/releases/%s/", alpineMirror, release, arch)
			releaseUrl := mirror + "latest-releases.yaml"
			wg.Add(1)
			go func() {
				defer wg.Done()
				page, err := capturePage(releaseUrl)
				if err != nil {
					errs <- Failure{Release: release, Arch: arch, Error: err}
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

	return waitForConfigs(ch, &wg), nil
}
