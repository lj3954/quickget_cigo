package os

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
)

const (
	almaMirror    = "https://repo.almalinux.org/almalinux/"
	almaReleaseRe = `<a href="([0-9]+)/"`
)

type Alma struct{}

func (Alma) Data() OSData {
	return OSData{
		Name:        "alma",
		PrettyName:  "AlmaLinux",
		Homepage:    "https://almalinux.org/",
		Description: "Community owned and governed, forever-free enterprise Linux distribution, focused on long-term stability, providing a robust production-grade platform. AlmaLinux OS is binary compatible with RHEL®.",
	}
}

func (Alma) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	releases, numReleases, err := getBasicReleases(almaMirror, almaReleaseRe, -1)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannelsWith(numReleases * len(x86_64_aarch64))
	isoRe := regexp.MustCompile(`<a href="(AlmaLinux-[0-9]+-latest-(?:x86_64|aarch64)-([^-]+).iso)">`)

	for release := range releases {
		for _, arch := range x86_64_aarch64 {
			mirror := fmt.Sprintf("%s%s/isos/%s/", almaMirror, release, arch)
			go func() {
				defer wg.Done()

				page, err := capturePage(mirror)
				if err != nil {
					errs <- Failure{Release: release, Arch: arch, Error: err}
					return
				}
				checksums, err := cs.Build(cs.Sha256Regex, mirror+"CHECKSUM")
				if err != nil {
					csErrs <- Failure{Release: release, Arch: arch, Error: err}
				}
				for _, match := range isoRe.FindAllStringSubmatch(page, -1) {
					if strings.HasSuffix(match[0], ".manifest") {
						continue
					}
					iso, edition := match[1], match[2]
					url := mirror + iso
					checksum := checksums[iso]
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
	}

	return waitForConfigs(ch, wg), nil
}
